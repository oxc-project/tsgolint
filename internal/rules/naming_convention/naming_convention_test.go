package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNamingConventionDefault(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Default config: camelCase for most things, PascalCase for types, camelCase|UPPER_CASE for variables
		{Code: "const myVar = 1;"},
		{Code: "let myVar = 1;"},
		{Code: "var myVar = 1;"},
		{Code: "const MY_VAR = 1;"},
		{Code: "function myFunc() {}"},
		{Code: "class MyClass {}"},
		{Code: "interface MyInterface {}"},
		{Code: "type MyType = {};"},
		{Code: "enum MyEnum {}"},
		{Code: "const _myVar = 1;"},
		{Code: "const myVar_ = 1;"},
		// Imports (default and namespace only - named imports are not matched)
		{Code: "import myDefault from 'module';"},
		{Code: "import MyDefault from 'module';"},
	}, []rule_tester.InvalidTestCase{
		// Variables with PascalCase should fail (default requires camelCase or UPPER_CASE)
		{
			Code: "const MyVar = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// Functions with PascalCase should fail
		{
			Code: "function MyFunc() {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// Classes with camelCase should fail
		{
			Code: "class myClass {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// Interfaces with camelCase should fail
		{
			Code: "interface myInterface {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// Type aliases with snake_case should fail
		{
			Code: "type my_type = {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		// Enums with camelCase should fail
		{
			Code: "enum myEnum {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionVariables(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const myVar = 1;", Options: opts},
		{Code: "let myVar = 1;", Options: opts},
		{Code: "var myVar = 1;", Options: opts},
		{Code: "const lower = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const snake_case = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "const UPPER_CASE = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "const StrictPascalCase = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionFunctions(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "function", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "function myFunc() {}", Options: opts},
		{Code: "function lower() {}", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "function MyFunc() {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "function UPPER_CASE() {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionClasses(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "class", Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class MyClass {}", Options: opts},
		{Code: "abstract class MyAbstractClass {}", Options: opts},
		{Code: "const ignored = class MyExpr {}", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "class myClass {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "class snake_case {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionInterfaces(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "interface", Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "interface MyInterface {}", Options: opts},
		{Code: "interface Pascal {}", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface myInterface {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionTypeAliases(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "typeAlias", Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "type MyType = {};", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "type my_type = {};",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionEnums(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "enum", Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "enum MyEnum {}", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "enum myEnum {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionEnumMembers(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "enumMember", Format: &[]string{"UPPER_CASE"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "enum MyEnum { MY_MEMBER = 1 }", Options: opts},
		{Code: "enum MyEnum { UPPER = 1 }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "enum MyEnum { myMember = 1 }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionTypeParameters(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "typeParameter", Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "interface Ignored<T> extends Array<T> {}", Options: opts},
		{Code: "function ignored<TParam>(x: TParam) { return x; }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "function ignored<t>(x: t) { return x; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionParameters(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "parameter", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "function fn(myParam: string) {}", Options: opts},
		{Code: "(function(myParam: string) {});", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "function fn(MyParam: string) {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionClassProperties(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "classProperty", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class Foo { myProp = 1; }", Options: opts},
		{Code: "class Foo { private myProp = 1; }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "class Foo { MyProp = 1; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "class Foo { UPPER_CASE = 1; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionClassMethods(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "classMethod", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class Foo { myMethod() {} }", Options: opts},
		{Code: "class Foo { private myMethod() {} }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "class Foo { MyMethod() {} }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionObjectLiteralProperties(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "objectLiteralProperty", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const obj = { myProp: 1 };", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const obj = { MY_PROP: 1 };",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionObjectLiteralMethods(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "objectLiteralMethod", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const obj = { myMethod() {} };", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const obj = { MyMethod() {} };",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionTypeProperties(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "typeProperty", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "interface Foo { myProp: string; }", Options: opts},
		{Code: "type Foo = { myProp: string; };", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface Foo { MY_PROP: string; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionTypeMethods(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "typeMethod", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "interface Foo { myMethod(): void; }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface Foo { MyMethod(): void; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionGetSetAccessors(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "classicAccessor", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class Foo { get myProp() { return 1; } }", Options: opts},
		{Code: "class Foo { set myProp(value: number) {} }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "class Foo { get MyProp() { return 1; } }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionLeadingUnderscore(t *testing.T) {
	t.Parallel()

	forbidOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("forbid")},
	}
	requireOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("require")},
	}
	allowOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow")},
	}
	requireDoubleOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("requireDouble")},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Forbid
		{Code: "const myVar = 1;", Options: forbidOpts},
		// Require
		{Code: "const _myVar = 1;", Options: requireOpts},
		// Allow
		{Code: "const myVar = 1;", Options: allowOpts},
		{Code: "const _myVar = 1;", Options: allowOpts},
		// RequireDouble
		{Code: "const __myVar = 1;", Options: requireDoubleOpts},
	}, []rule_tester.InvalidTestCase{
		// Forbid - underscore present
		{
			Code:    "const _myVar = 1;",
			Options: forbidOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUnderscore"}},
		},
		// Require - no underscore
		{
			Code:    "const myVar = 1;",
			Options: requireOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingUnderscore"}},
		},
		// RequireDouble - single underscore
		{
			Code:    "const _myVar = 1;",
			Options: requireDoubleOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingUnderscore"}},
		},
	})
}

func TestNamingConventionTrailingUnderscore(t *testing.T) {
	t.Parallel()

	forbidOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, TrailingUnderscore: strPtr("forbid")},
	}
	requireOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, TrailingUnderscore: strPtr("require")},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const myVar = 1;", Options: forbidOpts},
		{Code: "const myVar_ = 1;", Options: requireOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const myVar_ = 1;",
			Options: forbidOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUnderscore"}},
		},
		{
			Code:    "const myVar = 1;",
			Options: requireOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingUnderscore"}},
		},
	})
}

func TestNamingConventionPrefixSuffix(t *testing.T) {
	t.Parallel()

	prefixOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"PascalCase"}, Prefix: []string{"is", "has"}},
	}
	suffixOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, Suffix: []string{"Type", "Interface"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const isEnabled = true;", Options: prefixOpts},
		{Code: "const hasValue = true;", Options: prefixOpts},
		{Code: "const myVarType = 1;", Options: suffixOpts},
		{Code: "const myVarInterface = 1;", Options: suffixOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const Enabled = true;",
			Options: prefixOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingAffix"}},
		},
		{
			Code:    "const myVar = 1;",
			Options: suffixOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingAffix"}},
		},
	})
}

func TestNamingConventionCustomRegex(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, Custom: &MatchRegex{Match: true, Regex: "^x"}},
	}
	notMatchOpts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}, Custom: &MatchRegex{Match: false, Regex: "^unused_"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const xMyVar = 1;", Options: opts},
		{Code: "const myVar = 1;", Options: notMatchOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const myVar = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
		{
			Code:    "const unused_myVar = 1;",
			Options: notMatchOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
	})
}

func TestNamingConventionFilter(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{
			Selector: "variable",
			Format:   &[]string{"PascalCase"},
			Filter:   map[string]any{"regex": "^ignore", "match": false},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const MyVar = 1;", Options: opts},
		// Filtered out - doesn't match, so not checked
		{Code: "const ignoredVar = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const myVar = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionNullFormat(t *testing.T) {
	t.Parallel()

	// format: null means no format enforcement
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "variable", Format: nil},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Variable should be unchecked since format is null
		{Code: "const anything_goes = 1;", Options: opts},
		{Code: "const ALSO_FINE = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{
		// Function is checked by default selector (PascalCase)
		{
			Code:    "function myFunc() {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionMultipleFormats(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase", "UPPER_CASE"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const myVar = 1;", Options: opts},
		{Code: "const MY_VAR = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const MyVar = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionModifiers(t *testing.T) {
	t.Parallel()

	// Static class property
	staticOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "typeLike", Format: &[]string{"PascalCase"}},
		{Selector: "classProperty", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"static"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class MyClass {
  static MY_PROP = 1;
  myProp = 2;
}
`, Options: staticOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class MyClass {
  static myProp = 1;
}
`,
			Options: staticOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionReadonlyModifier(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "typeLike", Format: &[]string{"PascalCase"}},
		{Selector: "classProperty", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"static", "readonly"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Ignored {
  private static readonly SOME_NAME = 1;
  ignoredDueToModifiers = 1;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionDestructured(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Modifiers: []string{"destructured"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const { some_name } = {};
const IgnoredDueToModifiers = 1;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionExported(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "variable", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const camelCaseVar = 1;
export const PascalCaseVar = 1;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
export const camelCaseVar = 1;
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionMetaSelectors(t *testing.T) {
	t.Parallel()

	// typeLike meta selector
	typeOpts := []NamingConventionOption{
		{Selector: "typeLike", Format: &[]string{"PascalCase"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class MyClass {}", Options: typeOpts},
		{Code: "interface MyInterface {}", Options: typeOpts},
		{Code: "type MyType = {};", Options: typeOpts},
		{Code: "enum MyEnum {}", Options: typeOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "class myClass {}",
			Options: typeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "interface myInterface {}",
			Options: typeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "type myType = {};",
			Options: typeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "enum myEnum {}",
			Options: typeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionSpecificity(t *testing.T) {
	t.Parallel()

	// More specific selector overrides less specific
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "property", Format: &[]string{"PascalCase"}},
		{Selector: "method", Format: &[]string{"PascalCase"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const obj = {
  Foo: 42,
  Bar() { return 42; },
};
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionRequiresQuotes(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}},
		{Selector: "default", Format: nil, Modifiers: []string{"requiresQuotes"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const ignored = {
  'a a': 1,
  'b b'() {},
};
class ignored_class {
  'a a' = 1;
  'b b'() {}
}
enum ignored_enum {
  'a a',
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionHashPrivate(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: "memberLike", Format: &[]string{"snake_case"}, Modifiers: []string{"#private"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Foo {
  private someAttribute = 1;
  #some_attribute = 1;
  private someMethod() {}
  #some_method() {}
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionAsyncModifier(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const async_bar = async () => {};", Options: opts},
		{Code: "const myVar = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionImports(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "import", Format: &[]string{"PascalCase"}},
		{Selector: "import", Format: &[]string{"camelCase"}, Modifiers: []string{"default"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "import * as FooBar from 'module';", Options: opts},
		{Code: "import fooBar from 'module';", Options: opts},
		// Named imports are not matched by the import selector per spec
		{Code: "import { anything_goes } from 'module';", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "import * as foo_bar from 'module';",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionOverrideModifier(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: "memberLike", Format: &[]string{"snake_case"}, Modifiers: []string{"override"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Foo extends Object {
  public someAttribute = 1;
  public override some_attribute_override = 1;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionParameterProperty(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "parameterProperty", Format: &[]string{"snake_case"}, Modifiers: []string{"readonly"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Ignored {
  constructor(
    private readonly some_name: string,
    IgnoredDueToModifiers: string,
  ) {}
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionMethodSignature(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "typeMethod", Format: &[]string{"PascalCase"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "interface Foo { MyMethod(): void; }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface Foo { myMethod(): void; }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionPropertySignatureReadonly(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "typeProperty", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"readonly"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
type Ignored = {
  ignored_due_to_modifiers: string;
  readonly FOO: string;
};
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionMultipleSelectorArray(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: []string{"variable", "function"}, Format: &[]string{"snake_case"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const my_var = 1;", Options: opts},
		{Code: "function my_func() {}", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const myVar = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "function myFunc() {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionAbstractClass(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "class", Format: &[]string{"snake_case"}, Modifiers: []string{"abstract"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
abstract class some_name {}
class IgnoredDueToModifier {}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionConstEnum(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "typeLike", Format: &[]string{"PascalCase"}},
		{Selector: "enum", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"const"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const enum MY_CONST_ENUM {}
enum RegularEnum {}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const enum notUpperCase {}
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionStaticAccessor(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "accessor", Format: &[]string{"snake_case"}, Modifiers: []string{"private", "static"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Ignored {
  private static get some_name() { return 1; }
  get IgnoredDueToModifiers() { return 1; }
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionStaticMethod(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "classMethod", Format: &[]string{"snake_case"}, Modifiers: []string{"static"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Ignored {
  private static some_name() {}
  IgnoredDueToModifiers() {}
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionAllFormats(t *testing.T) {
	t.Parallel()

	// Test each format independently
	camelOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"camelCase"}}}
	strictCamelOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"strictCamelCase"}}}
	pascalOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"PascalCase"}}}
	strictPascalOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"StrictPascalCase"}}}
	snakeOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"snake_case"}}}
	upperOpts := []NamingConventionOption{{Selector: "variable", Format: &[]string{"UPPER_CASE"}}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// camelCase
		{Code: "const strictCamelCase = 1;", Options: camelOpts},
		{Code: "const lower = 1;", Options: camelOpts},
		{Code: "const camelCaseUNSTRICT = 1;", Options: camelOpts},
		// strictCamelCase
		{Code: "const strictCamelCase = 1;", Options: strictCamelOpts},
		{Code: "const lower = 1;", Options: strictCamelOpts},
		// PascalCase
		{Code: "const StrictPascalCase = 1;", Options: pascalOpts},
		{Code: "const Pascal = 1;", Options: pascalOpts},
		{Code: "const UPPER = 1;", Options: pascalOpts},
		{Code: "const PascalCaseUNSTRICT = 1;", Options: pascalOpts},
		// StrictPascalCase
		{Code: "const StrictPascalCase = 1;", Options: strictPascalOpts},
		{Code: "const Pascal = 1;", Options: strictPascalOpts},
		// snake_case
		{Code: "const snake_case = 1;", Options: snakeOpts},
		{Code: "const lower = 1;", Options: snakeOpts},
		// UPPER_CASE
		{Code: "const UPPER_CASE = 1;", Options: upperOpts},
		{Code: "const UPPER = 1;", Options: upperOpts},
	}, []rule_tester.InvalidTestCase{
		// camelCase rejects
		{Code: "const snake_case = 1;", Options: camelOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const UPPER_CASE = 1;", Options: camelOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const StrictPascalCase = 1;", Options: camelOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		// strictCamelCase rejects
		{Code: "const camelCaseUNSTRICT = 1;", Options: strictCamelOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const UPPER = 1;", Options: strictCamelOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		// PascalCase rejects
		{Code: "const snake_case = 1;", Options: pascalOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const strictCamelCase = 1;", Options: pascalOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		// StrictPascalCase rejects
		{Code: "const PascalCaseUNSTRICT = 1;", Options: strictPascalOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const UPPER = 1;", Options: strictPascalOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		// snake_case rejects
		{Code: "const UPPER_CASE = 1;", Options: snakeOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const strictCamelCase = 1;", Options: snakeOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const StrictPascalCase = 1;", Options: snakeOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		// UPPER_CASE rejects
		{Code: "const lower = 1;", Options: upperOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const snake_case = 1;", Options: upperOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
		{Code: "const strictCamelCase = 1;", Options: upperOpts, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}}},
	})
}

func TestNamingConventionFunctionValuedPropertyAsMethod(t *testing.T) {
	t.Parallel()

	// classMethod selector should match properties with function/arrow expression values
	classOpts := []NamingConventionOption{
		{Selector: "classProperty", Format: &[]string{"camelCase"}},
		{Selector: "classMethod", Format: &[]string{"PascalCase"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Arrow function property should be treated as classMethod (PascalCase)
		{Code: "class Foo { MyMethod = () => {}; }", Options: classOpts},
		// Function expression property should be treated as classMethod (PascalCase)
		{Code: "class Foo { MyMethod = function() {}; }", Options: classOpts},
		// Regular property should be treated as classProperty (camelCase)
		{Code: "class Foo { myProp = 1; }", Options: classOpts},
	}, []rule_tester.InvalidTestCase{
		// Arrow function property with camelCase should fail (expects PascalCase for classMethod)
		{
			Code:    "class Foo { myMethod = () => {}; }",
			Options: classOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// Regular property with PascalCase should fail (expects camelCase for classProperty)
		{
			Code:    "class Foo { MyProp = 1; }",
			Options: classOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionObjectLiteralFunctionValuedPropertyAsMethod(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "objectLiteralProperty", Format: &[]string{"camelCase"}},
		{Selector: "objectLiteralMethod", Format: &[]string{"PascalCase"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Arrow function property should be treated as objectLiteralMethod (PascalCase)
		{Code: "const obj = { MyMethod: () => {} };", Options: opts},
		// Function expression property should be treated as objectLiteralMethod (PascalCase)
		{Code: "const obj = { MyMethod: function() {} };", Options: opts},
		// Regular property should be treated as objectLiteralProperty (camelCase)
		{Code: "const obj = { myProp: 1 };", Options: opts},
	}, []rule_tester.InvalidTestCase{
		// Arrow function property with camelCase should fail (expects PascalCase for objectLiteralMethod)
		{
			Code:    "const obj = { myMethod: () => {} };",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionEnumMemberPublicModifier(t *testing.T) {
	t.Parallel()

	// memberLike with public modifier should match enum members
	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"public"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "enum MyEnum { MY_MEMBER = 1 }", Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "enum MyEnum { myMember = 1 }",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionAsyncFunctionValuedProperty(t *testing.T) {
	t.Parallel()

	opts := []NamingConventionOption{
		{Selector: "classMethod", Format: &[]string{"camelCase"}},
		{Selector: "classMethod", Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Async arrow function property should be treated as async classMethod
		{Code: "class Foo { async_method = async () => {}; }", Options: opts},
		// Non-async arrow function property should be regular classMethod
		{Code: "class Foo { myMethod = () => {}; }", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionDefaultSnakeCase(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		// make sure we handle no options and apply defaults
		{
			Code:   "const x_x = 1;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// make sure we handle empty options and apply defaults
		{
			Code:    "const x_x = 1;",
			Options: []NamingConventionOption{},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionFilterMatchTrue(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// filter match:false - child_process is excluded from check
		{Code: "const child_process = require('child_process');", Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"camelCase"}, Filter: map[string]any{"regex": "child_process", "match": false}},
		}},
	}, []rule_tester.InvalidTestCase{
		// filter match:true - child_process IS checked and fails camelCase
		{
			Code: "const child_process = require('child_process');",
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"camelCase"}, Filter: map[string]any{"regex": "child_process", "match": true}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionCustomRegexUnused(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}, LeadingUnderscore: strPtr("allow"), Custom: &MatchRegex{Match: false, Regex: "^unused_\\w"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code:    "let unused_foo = 'a';",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
		{
			Code:    "const _unused_foo = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
	})
}

func TestNamingConventionCustomRegexTypeLike(t *testing.T) {
	t.Parallel()
	typeLikeOpts := []NamingConventionOption{
		{Selector: "typeLike", Format: &[]string{"PascalCase"}, Custom: &MatchRegex{Match: false, Regex: "^I[A-Z]"}},
	}
	funcOpts := []NamingConventionOption{
		{Selector: "function", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow"), Custom: &MatchRegex{Match: true, Regex: "function"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface IFoo {}",
			Options: typeLikeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
		{
			Code:    "class IBar {}",
			Options: typeLikeOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
		{
			Code:    "function fooBar() {}",
			Options: funcOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
	})
}

func TestNamingConventionCustomRegexComprehensiveValid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow"), Custom: &MatchRegex{Match: false, Regex: "^unused_\\w"}},
		{Selector: "typeLike", Format: &[]string{"PascalCase"}, Custom: &MatchRegex{Match: false, Regex: "^I[A-Z]"}},
		{Selector: "function", Format: &[]string{"snake_case"}, LeadingUnderscore: strPtr("allow"), Custom: &MatchRegex{Match: true, Regex: "_function_"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "let foo = 'a';", Options: opts},
		{Code: "const _foo = 1;", Options: opts},
		{Code: "interface Foo {}", Options: opts},
		{Code: "class Bar {}", Options: opts},
		{Code: "function foo_function_bar() {}", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionCustomRegexArraySelector(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"default", "typeLike", "function"}, Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow"), Custom: &MatchRegex{Match: false, Regex: "^unused_\\w"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "let foo = 'a';", Options: opts},
		{Code: "const _foo = 1;", Options: opts},
		{Code: "interface foo {}", Options: opts},
		{Code: "class bar {}", Options: opts},
		{Code: "function fooFunctionBar() {}", Options: opts},
		{Code: "function _fooFunctionBar() {}", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionArraySelectorFormatErrors(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"variable", "function"}, Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow")},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code:    "let unused_foo = 'a';",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		{
			Code:    "const _unused_foo = 1;",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormatTrimmed"}},
		},
		{
			Code:    "function foo_bar() {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionCustomRegexMultipleSelectors(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"class", "interface"}, Format: &[]string{"PascalCase"}, Custom: &MatchRegex{Match: false, Regex: "^I[A-Z]"}},
	}
	multiOpts := []NamingConventionOption{
		{Selector: []string{"variable", "function"}, Format: &[]string{"camelCase"}, LeadingUnderscore: strPtr("allow")},
		{Selector: []string{"class", "interface"}, Format: &[]string{"PascalCase"}, Custom: &MatchRegex{Match: false, Regex: "^I[A-Z]"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code:    "interface IFoo {}",
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
		{
			Code:    "class IBar {}",
			Options: multiOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "satisfyCustom"}},
		},
	})
}

func TestNamingConventionFilterExcludePattern(t *testing.T) {
	t.Parallel()
	dashFilterOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"strictCamelCase"}, Filter: map[string]any{"regex": "-", "match": false}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const foo = { 'Property-Name': 'asdf' };", Options: dashFilterOpts},
		{Code: "const foo = { 'Property-Name': 'asdf' };", Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"strictCamelCase"}, Filter: map[string]any{"regex": "^(Property-Name)$", "match": false}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "const foo = { 'Property Name': 'asdf' };",
			Options: dashFilterOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionFilterShorthand(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"camelCase", "UPPER_CASE"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Filter: "child_process"},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const child_process = require('child_process');", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionPrivateReadonlyFormat(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "class foo { private fooBoo: number; }", Options: []NamingConventionOption{
			{Selector: []string{"property", "accessor"}, Format: &[]string{"camelCase"}, Modifiers: []string{"private"}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "class foo { private readonly fooBar: boolean; }",
			Options: []NamingConventionOption{
				{Selector: []string{"property", "accessor"}, Format: &[]string{"PascalCase"}, Modifiers: []string{"private", "readonly"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionPropertyVsVariableSelector(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "property", Format: &[]string{"PascalCase"}},
		{Selector: "variable", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class SomeClass {
  static OtherConstant = 'hello';
}
export const { OtherConstant: otherConstant } = SomeClass;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class SomeClass {
  static otherConstant = 'hello';
}
export const { otherConstant } = SomeClass;
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionDeclareClassParameter(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "parameter", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare class Foo {
  Bar(Baz: string): void;
}
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionExportedComprehensive(t *testing.T) {
	t.Parallel()
	invalidOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}},
		{Selector: "variable", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
		{Selector: "function", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
		{Selector: "class", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
		{Selector: "interface", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
		{Selector: "typeAlias", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
		{Selector: "enum", Format: &[]string{"camelCase"}, Modifiers: []string{"exported"}},
	}
	validOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "variable", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
		{Selector: "function", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
		{Selector: "class", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
		{Selector: "interface", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
		{Selector: "typeAlias", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
		{Selector: "enum", Format: &[]string{"PascalCase"}, Modifiers: []string{"exported"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const camelCaseVar = 1;
enum camelCaseEnum {}
class camelCaseClass {}
function camelCaseFunction() {}
interface camelCaseInterface {}
type camelCaseType = {};
export const PascalCaseVar = 1;
export enum PascalCaseEnum {}
export class PascalCaseClass {}
export function PascalCaseFunction() {}
export interface PascalCaseInterface {}
export type PascalCaseType = {};
`, Options: validOpts},
		{Code: `
const camelCaseVar = 1;
enum camelCaseEnum {}
class camelCaseClass {}
function camelCaseFunction() {}
interface camelCaseInterface {}
type camelCaseType = {};
const PascalCaseVar = 1;
enum PascalCaseEnum {}
class PascalCaseClass {}
function PascalCaseFunction() {}
interface PascalCaseInterface {}
type PascalCaseType = {};
export {
  PascalCaseVar,
  PascalCaseEnum,
  PascalCaseClass,
  PascalCaseFunction,
  PascalCaseInterface,
  PascalCaseType,
};
`, Options: validOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
export const PascalCaseVar = 1;
export enum PascalCaseEnum {}
export class PascalCaseClass {}
export function PascalCaseFunction() {}
export interface PascalCaseInterface {}
export type PascalCaseType = {};
`,
			Options: invalidOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
		{
			Code: `
const PascalCaseVar = 1;
enum PascalCaseEnum {}
class PascalCaseClass {}
function PascalCaseFunction() {}
interface PascalCaseInterface {}
type PascalCaseType = {};
export {
  PascalCaseVar,
  PascalCaseEnum,
  PascalCaseClass,
  PascalCaseFunction,
  PascalCaseInterface,
  PascalCaseType,
};
`,
			Options: invalidOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionGlobalModifier(t *testing.T) {
	t.Parallel()
	validOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "variable", Format: &[]string{"PascalCase"}, Modifiers: []string{"global"}},
		{Selector: "function", Format: &[]string{"PascalCase"}, Modifiers: []string{"global"}},
	}
	invalidOpts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}},
		{Selector: "variable", Format: &[]string{"camelCase"}, Modifiers: []string{"global"}},
		{Selector: "function", Format: &[]string{"camelCase"}, Modifiers: []string{"global"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
{
  const camelCaseVar = 1;
  function camelCaseFunction() {}
  declare function camelCaseDeclaredFunction();
}
const PascalCaseVar = 1;
function PascalCaseFunction() {}
declare function PascalCaseDeclaredFunction();
`, Options: validOpts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const PascalCaseVar = 1;
function PascalCaseFunction() {}
declare function PascalCaseDeclaredFunction();
`,
			Options: invalidOpts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionDestructuredVariableInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const { some_name1 } = {};
const { ignore: IgnoredDueToModifiers1 } = {};
const { some_name2 = 2 } = {};
const IgnoredDueToModifiers2 = 1;
`, Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"PascalCase"}},
			{Selector: "variable", Format: &[]string{"snake_case"}, Modifiers: []string{"destructured"}},
		}},
		{Code: `
const { some_name1 } = {};
const { ignore: IgnoredDueToModifiers1 } = {};
const { some_name2 = 2 } = {};
const IgnoredDueToModifiers2 = 1;
`, Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"PascalCase"}},
			{Selector: "variable", Format: nil, Modifiers: []string{"destructured"}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const { some_name1 } = {};
const { some_name2 = 2 } = {};
const { ignored: IgnoredDueToModifiers1 } = {};
const { ignored: IgnoredDueToModifiers2 = 3 } = {};
const IgnoredDueToModifiers3 = 1;
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "variable", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"destructured"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionDestructuredParameterInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
export function Foo(
  { aName },
  { anotherName = 1 },
  { ignored: IgnoredDueToModifiers1 },
  { ignored: IgnoredDueToModifiers1 = 2 },
  IgnoredDueToModifiers2,
) {}
`, Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"PascalCase"}},
			{Selector: "parameter", Format: &[]string{"camelCase"}, Modifiers: []string{"destructured"}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
export function Foo(
  { aName },
  { anotherName = 1 },
  { ignored: IgnoredDueToModifiers1 },
  { ignored: IgnoredDueToModifiers1 = 2 },
  IgnoredDueToModifiers2,
) {}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "parameter", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"destructured"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionDestructuringHoles(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const match = 'test'.match(/test/);\nconst [, key, value] = match;", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionEmptyFormatArray(t *testing.T) {
	t.Parallel()
	emptyFormat := []string{}
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"camelCase"}},
		{Selector: "variable", Format: &emptyFormat},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "const snake_case = 1;", Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionStaticReadonlyInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class Ignored {
  private static abstract readonly some_name;
  IgnoredDueToModifiers = 1;
}
`, Options: []NamingConventionOption{
			{Selector: "default", Format: &[]string{"PascalCase"}},
			{Selector: "classProperty", Format: &[]string{"snake_case"}, Modifiers: []string{"static", "readonly"}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Ignored {
  private static abstract readonly some_name;
  IgnoredDueToModifiers = 1;
}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "classProperty", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"static", "readonly"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionParameterPropertyReadonlyInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Ignored {
  constructor(
    private readonly some_name,
    IgnoredDueToModifiers,
  ) {}
}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "parameterProperty", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"readonly"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionClassMethodStaticInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Ignored {
  private static some_name() {}
  IgnoredDueToModifiers() {}
}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "classMethod", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"static"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionAccessorPrivateStaticInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Ignored {
  private static get some_name() {}
  get IgnoredDueToModifiers() {}
}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "accessor", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"private", "static"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionAbstractClassInvalid(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
abstract class some_name {}
class IgnoredDueToModifier {}
`,
			Options: []NamingConventionOption{
				{Selector: "default", Format: &[]string{"PascalCase"}},
				{Selector: "class", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"abstract"}},
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionUnusedModifier(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"PascalCase"}},
		{Selector: "default", Format: &[]string{"snake_case"}, Modifiers: []string{"unused"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		// Non-exported items are unused → must be snake_case
		// Exported items are not unused → must be PascalCase
		// Referenced parameters/type params are not unused → must be PascalCase
		{Code: `
const unused_var = 1;
function unused_func(
  unused_param: string,
) {}
class unused_class {}
interface unused_interface {}
type unused_type<
  unused_typeparam,
> = {};

export const UsedVar = 1;
export function UsedFunc(
  UsedParam: string,
) {
  return UsedParam;
}
export class UsedClass {}
export interface UsedInterface {}
export type UsedType<
  UsedTypeParam,
> = UsedTypeParam;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		// All non-exported PascalCase items are unused → must be snake_case → errors
		{
			Code: `
const UnusedVar = 1;
function UnusedFunc(
  UnusedParam: string,
) {}
class UnusedClass {}
interface UnusedInterface {}
type UnusedType<
  UnusedTypeParam,
> = {};
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionRequiresQuotesEnforced(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}},
		{Selector: "default", Format: &[]string{"PascalCase"}, Modifiers: []string{"requiresQuotes"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
const ignored1 = {
  'a a': 1,
  'b b'() {},
  get 'c c'() {
    return 1;
  },
  set 'd d'(value: string) {},
};
class ignored2 {
  'a a' = 1;
  'b b'() {}
  get 'c c'() {
    return 1;
  }
  set 'd d'(value: string) {}
}
interface ignored3 {
  'a a': 1;
  'b b'(): void;
}
type ignored4 = {
  'a a': 1;
  'b b'(): void;
};
enum ignored5 {
  'a a',
}
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionQuotedPropertyDefault(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
type Foo = {
  'foo     Bar': string;
  '': string;
  '0': string;
  'foo': string;
  'foo-bar': string;
  '#foo-bar': string;
};
interface Bar {
  'boo-----foo': string;
}
`,
			// 6, not 7 because 'foo' is valid camelCase
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionRequiresQuotesPerSelector(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"snake_case"}},
		{
			Selector:  []string{"classProperty", "objectLiteralProperty", "typeProperty", "classMethod", "objectLiteralMethod", "typeMethod", "accessor", "enumMember"},
			Format:    nil,
			Modifiers: []string{"requiresQuotes"},
		},
		{
			Selector: []string{"classProperty", "objectLiteralProperty", "typeProperty", "classMethod", "objectLiteralMethod", "typeMethod", "accessor", "enumMember"},
			Format:   &[]string{"PascalCase"},
		},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const ignored1 = {
  'a a': 1,
  'b b'() {},
  get 'c c'() {
    return 1;
  },
  set 'd d'(value: string) {},
};
class ignored2 {
  'a a' = 1;
  'b b'() {}
  get 'c c'() {
    return 1;
  }
  set 'd d'(value: string) {}
}
interface ignored3 {
  'a a': 1;
  'b b'(): void;
}
type ignored4 = {
  'a a': 1;
  'b b'(): void;
};
enum ignored5 {
  'a a',
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionAsyncClassMethodsComprehensive(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: "method", Format: &[]string{"PascalCase"}},
		{Selector: []string{"method", "objectLiteralMethod"}, Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const obj = {
  Bar() {
    return 42;
  },
  async async_bar() {
    return 42;
  },
};
class foo {
  public Bar() {
    return 42;
  }
  public async async_bar() {
    return 42;
  }
}
abstract class foo {
  public Bar() {
    return 42;
  }
  public async async_bar() {
    return 42;
  }
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class foo {
  public Bar() {
    return 42;
  }
  public async async_bar() {
    return 42;
  }
  public async asyncBar() {
    return 42;
  }
  public AsyncBar2 = async () => {
    return 42;
  };
  public AsyncBar3 = async function () {
    return 42;
  };
}
abstract class foo {
  public abstract Bar(): number;
  public abstract async async_bar(): number;
  public abstract async ASYNC_BAR(): number;
}
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionAsyncObjectLiteralMethodsComprehensive(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: "method", Format: &[]string{"PascalCase"}},
		{Selector: []string{"method", "objectLiteralMethod"}, Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
const obj = {
  Bar() {
    return 42;
  },
  async async_bar() {
    return 42;
  },
  async AsyncBar() {
    return 42;
  },
  AsyncBar2: async () => {
    return 42;
  },
  AsyncBar3: async function () {
    return 42;
  },
};
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionAsyncVariableLike(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variableLike", Format: &[]string{"camelCase"}},
		{Selector: []string{"variableLike"}, Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const async_bar1 = async () => {};
async function async_bar2() {}
const async_bar3 = async function async_bar4() {};
`, Options: []NamingConventionOption{
			{Selector: "memberLike", Format: &[]string{"camelCase"}},
			{Selector: "method", Format: &[]string{"PascalCase"}},
			{Selector: []string{"variable"}, Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
		}},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const syncbar1 = () => {};
function syncBar2() {}
const syncBar3 = function syncBar4() {};

const AsyncBar1 = async () => {};
const async_bar1 = async () => {};
const async_bar3 = async function async_bar4() {};
async function async_bar2() {}
const asyncBar5 = async function async_bar6() {};
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionAsyncFunctionDeclarations(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variableLike", Format: &[]string{"camelCase"}},
		{Selector: []string{"variableLike"}, Format: &[]string{"snake_case"}, Modifiers: []string{"async"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
const syncbar1 = () => {};
function syncBar2() {}
const syncBar3 = function syncBar4() {};

const async_bar1 = async () => {};
async function asyncBar2() {}
const async_bar3 = async function async_bar4() {};
async function async_bar2() {}
const async_bar3 = async function ASYNC_BAR4() {};
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionOverrideComprehensive(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: []string{"memberLike"}, Format: &[]string{"snake_case"}, Modifiers: []string{"override"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class foo extends bar {
  public someAttribute = 1;
  public override some_attribute_override = 1;
  public someMethod() {
    return 42;
  }
  public override some_method_override2() {
    return 42;
  }
}
abstract class foo extends bar {
  public abstract someAttribute: string;
  public abstract override some_attribute_override: string;
  public abstract someMethod(): string;
  public abstract override some_method_override2(): string;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		// Override property
		{
			Code: `
class foo extends bar {
  public someAttribute = 1;
  public override some_attribute_override = 1;
  public override someAttributeOverride = 1;
}
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// Override method
		{
			Code: `
class foo extends bar {
  public override some_method_override() {
    return 42;
  }
  public override someMethodOverride() {
    return 42;
  }
}
`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// Override accessors
		{
			Code: `
class foo extends bar {
  public get someGetter(): string;
  public override get some_getter_override(): string;
  public override get someGetterOverride(): string;
  public set someSetter(val: string);
  public override set some_setter_override(val: string);
  public override set someSetterOverride(val: string);
}
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionHashPrivateComprehensiveInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "memberLike", Format: &[]string{"camelCase"}},
		{Selector: []string{"memberLike"}, Format: &[]string{"snake_case"}, Modifiers: []string{"#private"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{
			Code: `
class foo {
  private firstPrivateField = 1;
  private first_private_field = 1;
  #secondPrivateField = 1;
  #second_private_field = 1;
}
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
		{
			Code: `
class foo {
  private firstPrivateMethod() {}
  private first_private_method() {}
  #secondPrivateMethod() {}
  #second_private_method() {}
}
`,
			Options: opts,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
	})
}

func TestNamingConventionImportModifiers(t *testing.T) {
	t.Parallel()
	invalidOpts := []NamingConventionOption{
		{Selector: []string{"import"}, Format: &[]string{"camelCase"}},
		{Selector: []string{"import"}, Format: &[]string{"PascalCase"}, Modifiers: []string{"namespace"}},
	}
	validOpts := []NamingConventionOption{
		{Selector: []string{"import"}, Format: &[]string{"PascalCase"}},
		{Selector: []string{"import"}, Format: &[]string{"camelCase"}, Modifiers: []string{"default"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: "import * as FooBar from 'foo_bar';", Options: validOpts},
		{Code: "import fooBar from 'foo_bar';", Options: validOpts},
		{Code: "import { default as fooBar } from 'foo_bar';", Options: validOpts},
		// Named imports are not matched by import selector
		{Code: "import { foo_bar } from 'foo_bar';", Options: validOpts},
	}, []rule_tester.InvalidTestCase{
		// Namespace import fails PascalCase
		{
			Code:    "import * as fooBar from 'foo_bar';",
			Options: invalidOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// Default import fails camelCase
		{
			Code:    "import FooBar from 'foo_bar';",
			Options: invalidOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
		// Default destructured import fails camelCase
		{
			Code:    "import { default as foo_bar } from 'foo_bar';",
			Options: invalidOpts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionImportEmoji(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"import"}, Format: &[]string{"PascalCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `import { "🍎" as Foo } from 'foo_bar';`, Options: opts},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    `import { "🍎" as foo } from 'foo_bar';`,
			Options: opts,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "doesNotMatchFormat"}},
		},
	})
}

func TestNamingConventionInterfaceFunctionPropertyAsTypeMethod(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "default", Format: &[]string{"UPPER_CASE"}},
		{Selector: "typeMethod", Format: &[]string{"PascalCase"}},
		{Selector: "typeProperty", Format: &[]string{"snake_case"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
interface SOME_INTERFACE {
  SomeMethod: () => void;
  some_property: string;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionPropertyVsVariableDestructuredRename(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "property", Format: &[]string{"PascalCase"}},
		{Selector: "variable", Format: &[]string{"camelCase"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class SomeClass {
  static OtherConstant = 'hello';
}
export const { OtherConstant: otherConstant } = SomeClass;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

// ===================== Type-aware tests =====================

func TestNamingConventionTypesMultipleInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"const"}, Prefix: []string{"any_"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Types: []string{"string"}, Prefix: []string{"string_"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Types: []string{"number"}, Prefix: []string{"number_"}},
		{Selector: "variable", Format: &[]string{"snake_case"}, Types: []string{"boolean"}, Prefix: []string{"boolean_"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: `
declare const any_camelCase01: any;
declare const any_camelCase02: any | null;
declare const any_camelCase03: any | null | undefined;
declare const string_camelCase01: string;
declare const string_camelCase02: string | null;
declare const string_camelCase03: string | null | undefined;
declare const string_camelCase04: 'a' | null | undefined;
declare const string_camelCase05: string | 'a' | null | undefined;
declare const number_camelCase06: number;
declare const number_camelCase07: number | null;
declare const number_camelCase08: number | null | undefined;
declare const number_camelCase09: 1 | null | undefined;
declare const number_camelCase10: number | 2 | null | undefined;
declare const boolean_camelCase11: boolean;
declare const boolean_camelCase12: boolean | null;
declare const boolean_camelCase13: boolean | null | undefined;
declare const boolean_camelCase14: true | null | undefined;
declare const boolean_camelCase15: false | null | undefined;
declare const boolean_camelCase16: true | false | null | undefined;
`, Options: opts, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
		}},
	})
}

func TestNamingConventionTypesFunctionInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"snake_case"}, Types: []string{"function"}, Prefix: []string{"function_"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: `
declare const function_camelCase1: () => void;
declare const function_camelCase2: (() => void) | null;
declare const function_camelCase3: (() => void) | null | undefined;
declare const function_camelCase4:
  | (() => void)
  | (() => string)
  | null
  | undefined;
`, Options: opts, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
		}},
	})
}

func TestNamingConventionTypesArrayInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"snake_case"}, Types: []string{"array"}, Prefix: []string{"array_"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: `
declare const array_camelCase1: Array<number>;
declare const array_camelCase2: ReadonlyArray<number> | null;
declare const array_camelCase3: number[] | null | undefined;
declare const array_camelCase4: readonly number[] | null | undefined;
declare const array_camelCase5:
  | number[]
  | (number | string)[]
  | null
  | undefined;
declare const array_camelCase6: [] | null | undefined;
declare const array_camelCase7: [number] | null | undefined;
declare const array_camelCase8:
  | readonly number[]
  | Array<string>
  | [boolean]
  | null
  | undefined;
`, Options: opts, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
		}},
	})
}

func TestNamingConventionTypesStringMultiSelectorInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"variable", "property", "parameter"}, Format: &[]string{"PascalCase"}, Types: []string{"string"}, Prefix: []string{"my", "My"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: `
const myfoo_bar = 'abcs';
function fun(myfoo: string) {}
class foo {
  Myfoo: string;
}
`, Options: opts, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
			{MessageId: "doesNotMatchFormatTrimmed"},
		}},
	})
}

func TestNamingConventionTypesStringFunctionSelectorInvalid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"variable", "function"}, Format: &[]string{"PascalCase"}, Types: []string{"string"}, Prefix: []string{"my", "My"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: `
function my_foo_bar() {}
`, Options: opts, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "doesNotMatchFormatTrimmed"},
		}},
	})
}

func TestNamingConventionTypesMultipleValid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: "variable", Format: &[]string{"UPPER_CASE"}, Modifiers: []string{"const"}, Prefix: []string{"ANY_"}},
		{Selector: "variable", Format: &[]string{"camelCase"}, Types: []string{"string"}, Prefix: []string{"string_"}},
		{Selector: "variable", Format: &[]string{"camelCase"}, Types: []string{"number"}, Prefix: []string{"number_"}},
		{Selector: "variable", Format: &[]string{"camelCase"}, Types: []string{"boolean"}, Prefix: []string{"boolean_"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
declare const ANY_UPPER_CASE: any;
declare const ANY_UPPER_CASE: any | null;
declare const ANY_UPPER_CASE: any | null | undefined;

declare const string_camelCase: string;
declare const string_camelCase: string | null;
declare const string_camelCase: string | null | undefined;
declare const string_camelCase: 'a' | null | undefined;
declare const string_camelCase: string | 'a' | null | undefined;

declare const number_camelCase: number;
declare const number_camelCase: number | null;
declare const number_camelCase: number | null | undefined;
declare const number_camelCase: 1 | null | undefined;
declare const number_camelCase: number | 2 | null | undefined;

declare const boolean_camelCase: boolean;
declare const boolean_camelCase: boolean | null;
declare const boolean_camelCase: boolean | null | undefined;
declare const boolean_camelCase: true | null | undefined;
declare const boolean_camelCase: false | null | undefined;
declare const boolean_camelCase: true | false | null | undefined;
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionTypesNumberPrefixValid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"variable", "parameter", "property", "accessor"}, Format: &[]string{"PascalCase"}, Types: []string{"number"}, Prefix: []string{"is", "should", "has", "can", "did", "will"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
let isFoo = 1;
class foo {
  shouldBoo: number;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionTypesBooleanPrivateReadonlyValid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"property", "accessor"}, Format: &[]string{"PascalCase"}, Modifiers: []string{"private", "readonly"}, Types: []string{"boolean"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
class foo {
  private readonly FooBoo: boolean;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}

func TestNamingConventionTypesNumberPrefixMultiRuleValid(t *testing.T) {
	t.Parallel()
	opts := []NamingConventionOption{
		{Selector: []string{"property", "accessor"}, Format: &[]string{"StrictPascalCase"}, Modifiers: []string{"private"}, Prefix: []string{"Van"}},
		{Selector: []string{"variable", "parameter"}, Format: &[]string{"camelCase"}, Types: []string{"number"}, Prefix: []string{"is", "good"}},
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, []rule_tester.ValidTestCase{
		{Code: `
const isfooBar = 1;
function fun(goodfunFoo: number) {}
class foo {
  private VanFooBar: number;
}
`, Options: opts},
	}, []rule_tester.InvalidTestCase{})
}
