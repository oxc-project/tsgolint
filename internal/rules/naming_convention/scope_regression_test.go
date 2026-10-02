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

func TestNamingConventionLogicalAssignments(t *testing.T) {
	options := modifierTestOptions("variable", "unused")
	valid := []rule_tester.ValidTestCase{
		{Code: `let BadName = 0; BadName += 1;`, Options: options},
		{Code: `let BadName = 0; BadName *= 1;`, Options: options},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, operator := range []string{"&&=", "||=", "??="} {
		code := "let BadName = 0; BadName " + operator + " 1;"
		expected := modifierFailure(code)
		expected.Message = "Variable name `BadName` must match one of the following formats: snake_case"
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Options: options, Errors: []rule_tester.InvalidTestCaseError{expected}})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}

func TestNamingConventionBodylessSignatureParameters(t *testing.T) {
	options := modifierTestOptions("parameter", "unused")
	message := "Parameter name `BadName` must match one of the following formats: snake_case"
	invalid := []rule_tester.InvalidTestCase{
		{
			Code:    `declare function f(BadName: string): void;`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 20, EndLine: 1, EndColumn: 35}},
		},
		{
			Code:    `declare function f({ BadName }: { BadName: string }): void;`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 22, EndLine: 1, EndColumn: 29}},
		},
		{
			Code:    `declare class Example { method(BadName: number): void; }`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 32, EndLine: 1, EndColumn: 47}},
		},
		{
			Code:    `declare class Example { method({ BadName }: { BadName: number }): void; }`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 34, EndLine: 1, EndColumn: 41}},
		},
		{
			Code:    `declare class Example { constructor(BadName: string); }`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 37, EndLine: 1, EndColumn: 52}},
		},
		{
			Code:    `declare class Example { constructor({ BadName }: { BadName: string }); }`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 39, EndLine: 1, EndColumn: 46}},
		},
		{
			Code:    `abstract class Example { abstract method(BadName: number): void; }`,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat", Message: message, Line: 1, Column: 42, EndLine: 1, EndColumn: 57}},
		},
	}
	valid := []rule_tester.ValidTestCase{
		{Code: `function f(BadName: string) {}`, Options: options},
		{Code: `function f({ BadName }: { BadName: string }) {}`, Options: options},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}

func TestNamingConventionTypePredicateReferences(t *testing.T) {
	options := append(modifierTestOptions("typeAlias", "unused"), modifierTestOptions("parameter", "unused")...)
	qualifiedOptions := append(NamingConventionOptions{}, options...)
	qualifiedOptions = append(qualifiedOptions, NamingConventionOptions{{Selector: NamingSelector{"typeAlias"}, Modifiers: []string{"exported"}, Format: []string{"PascalCase"}}}...)
	valid := []rule_tester.ValidTestCase{
		{Code: `function isBad(ValueName: unknown): ValueName is string { return true; }`, Options: modifierTestOptions("parameter", "unused")},
		{
			Code:    `namespace Types { export type BadName = {}; function isBad(ValueName: unknown): ValueName is Types.BadName { return true; } }`,
			Options: qualifiedOptions,
		},
	}
	codes := []string{
		`type BadName = {}; function isBad(ValueName: unknown): ValueName is BadName { return true; }`,
		`type BadName = {}; function isBad(ValueName: unknown): ValueName is Array<BadName> { return true; }`,
		`type BadName = {}; function isBad(ValueName: unknown): ValueName is Promise<Array<BadName>> { return true; }`,
	}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(codes))
	for _, code := range codes {
		expected := modifierFailure(code)
		expected.Message = "Type Alias name `BadName` must match one of the following formats: snake_case"
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:    code,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{expected},
		})
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}

func TestNamingConventionForInOfSingleReturnUsage(t *testing.T) {
	options := modifierTestOptions("variable", "unused")
	valid := []rule_tester.ValidTestCase{
		{Code: `function f(items: Record<string, number>) { for (const BadName in items) { void 0; return true; } }`, Options: options},
		{Code: `function f(items: Record<string, number>) { for (const BadName in items) {} }`, Options: options},
		{Code: `function f(items: [number, number][]) { let BadName, OtherName; for ([BadName, OtherName] of items) return true; }`, Options: options},
		{Code: `function f(items: unknown[]) { for (const [] of items) return true; }`, Options: options},
	}
	codes := []string{
		`function f(items: Record<string, number>) { for (const BadName in items) return true; }`,
		`function f(items: [number, number][]) { for (const [BadName, OtherName] of items) return true; }`,
		`function f(items: Record<string, number>) { let BadName; for (BadName in items) return true; }`,
		`function f(items: Record<string, number>) { for (const BadName in items) { return true; } }`,
	}
	invalid := make([]rule_tester.InvalidTestCase, 0, len(codes))
	for _, code := range codes {
		expected := modifierFailure(code)
		expected.Message = "Variable name `BadName` must match one of the following formats: snake_case"
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:    code,
			Options: options,
			Errors:  []rule_tester.InvalidTestCaseError{expected},
		})
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
}
