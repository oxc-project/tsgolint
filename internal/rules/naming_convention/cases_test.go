package naming_convention

import "testing"

// Ports of the upstream per-selector dynamic test case files.
// Source: https://github.com/typescript-eslint/typescript-eslint/tree/main/packages/eslint-plugin/tests/rules/naming-convention/cases
// Each test function corresponds 1:1 to the upstream file named in its comment.

// cases/accessor.test.ts
func TestNamingConventionCasesAccessor(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored { accessor % = 10; }`,
				`class Ignored { accessor #% = 10; }`,
				`class Ignored { static accessor % = 10; }`,
				`class Ignored { static accessor #% = 10; }`,
				`class Ignored { private accessor % = 10; }`,
				`class Ignored { private static accessor % = 10; }`,
				`class Ignored { override accessor % = 10; }`,
				`class Ignored { accessor "%" = 10; }`,
				`class Ignored { protected accessor % = 10; }`,
				`class Ignored { public accessor % = 10; }`,
				`class Ignored { abstract accessor %; }`,
				`const ignored = { get %() {} };`,
				`const ignored = { set "%"(ignored) {} };`,
				`class Ignored { private get %() {} }`,
				`class Ignored { private set "%"(ignored) {} }`,
				`class Ignored { private static get %() {} }`,
				`class Ignored { static get #%() {} }`,
			},
			options: NamingConventionOption{
				Selector: "accessor",
			},
		},
	})
}

// cases/autoAccessor.test.ts
func TestNamingConventionCasesAutoAccessor(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored { accessor % = 10; }`,
				`class Ignored { accessor #% = 10; }`,
				`class Ignored { static accessor % = 10; }`,
				`class Ignored { static accessor #% = 10; }`,
				`class Ignored { private accessor % = 10; }`,
				`class Ignored { private static accessor % = 10; }`,
				`class Ignored { override accessor % = 10; }`,
				`class Ignored { accessor "%" = 10; }`,
				`class Ignored { protected accessor % = 10; }`,
				`class Ignored { public accessor % = 10; }`,
				`class Ignored { abstract accessor %; }`,
			},
			options: NamingConventionOption{
				Selector: "autoAccessor",
			},
		},
	})
}

// cases/class.test.ts
func TestNamingConventionCasesClass(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class % {}`,
				`abstract class % {}`,
				`const ignored = class % {}`,
			},
			options: NamingConventionOption{
				Selector: "class",
			},
		},
	})
}

// cases/classicAccessor.test.ts
func TestNamingConventionCasesClassicAccessor(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`const ignored = { get %() {} };`,
				`const ignored = { set "%"(ignored) {} };`,
				`class Ignored { private get %() {} }`,
				`class Ignored { private set "%"(ignored) {} }`,
				`class Ignored { private static get %() {} }`,
				`class Ignored { static get #%() {} }`,
				`abstract class Ignored { abstract get %(): number }`,
				`abstract class Ignored { abstract set %(ignored: number) }`,
			},
			options: NamingConventionOption{
				Selector: "classicAccessor",
			},
		},
	})
}

// cases/default.test.ts
func TestNamingConventionCasesDefault(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`const % = 1;`,
				`function % () {}`,
				`(function (%) {});`,
				`class Ignored { constructor(private %) {} }`,
				`const ignored = { % };`,
				`interface Ignored { %: string }`,
				`type Ignored = { %: string }`,
				`class Ignored { private % = 1 }`,
				`class Ignored { #% = 1 }`,
				`class Ignored { constructor(private %) {} }`,
				`class Ignored { #%() {} }`,
				`class Ignored { private %() {} }`,
				`const ignored = { %() {} };`,
				`class Ignored { private get %() {} }`,
				`enum Ignored { % }`,
				`abstract class % {}`,
				`interface % { }`,
				`type % = { };`,
				`enum % {}`,
				`interface Ignored<%> extends Ignored<string> {}`,
			},
			options: NamingConventionOption{
				Filter:   "[iI]gnored",
				Selector: "default",
			},
		},
	})
}

// cases/enum.test.ts
func TestNamingConventionCasesEnum(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`enum % {}`,
			},
			options: NamingConventionOption{
				Selector: "enum",
			},
		},
	})
}

// cases/enumMember.test.ts
func TestNamingConventionCasesEnumMember(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`enum Ignored { % }`,
				`enum Ignored { "%" }`,
			},
			options: NamingConventionOption{
				Selector: "enumMember",
			},
		},
	})
}

// cases/function.test.ts
func TestNamingConventionCasesFunction(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`function % () {}`,
				`(function % () {});`,
				`declare function % ();`,
			},
			options: NamingConventionOption{
				Selector: "function",
			},
		},
	})
}

// cases/interface.test.ts
func TestNamingConventionCasesInterface(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`interface % {}`,
			},
			options: NamingConventionOption{
				Selector: "interface",
			},
		},
	})
}

// cases/method.test.ts
func TestNamingConventionCasesMethod(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored { private %() {} }`,
				`class Ignored { private "%"() {} }`,
				`class Ignored { private async %() {} }`,
				`class Ignored { private static %() {} }`,
				`class Ignored { private static async %() {} }`,
				`class Ignored { private % = () => {} }`,
				`class Ignored { abstract %() }`,
				`class Ignored { #%() }`,
				`class Ignored { static #%() }`,
			},
			options: NamingConventionOption{
				Selector: "classMethod",
			},
		},
		{
			code: []string{
				`const ignored = { %() {} };`,
				`const ignored = { "%"() {} };`,
				`const ignored = { %: () => {} };`,
			},
			options: NamingConventionOption{
				Selector: "objectLiteralMethod",
			},
		},
		{
			code: []string{
				`interface Ignored { %(): string }`,
				`interface Ignored { "%"(): string }`,
				`interface Ignored { %: () => string }`,
				`interface Ignored { "%": () => string }`,
				`type Ignored = { %(): string }`,
				`type Ignored = { "%"(): string }`,
				`type Ignored = { %: () => string }`,
				`type Ignored = { "%": () => string }`,
			},
			options: NamingConventionOption{
				Selector: "typeMethod",
			},
		},
	})
}

// cases/parameter.test.ts
func TestNamingConventionCasesParameter(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`function ignored(%) {}`,
				`(function (%) {});`,
				`declare function ignored(%);`,
				`function ignored({%}) {}`,
				`function ignored(...%) {}`,
				`function ignored({% = 1}) {}`,
				`function ignored({...%}) {}`,
				`function ignored([%]) {}`,
				`function ignored([% = 1]) {}`,
				`function ignored([...%]) {}`,
			},
			options: NamingConventionOption{
				Selector: "parameter",
			},
		},
	})
}

// cases/parameterProperty.test.ts
func TestNamingConventionCasesParameterProperty(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored { constructor(private %) {} }`,
				`class Ignored { constructor(readonly %) {} }`,
				`class Ignored { constructor(private readonly %) {} }`,
			},
			options: NamingConventionOption{
				Selector: "parameterProperty",
			},
		},
		{
			code: []string{
				`class Ignored { constructor(private readonly %) {} }`,
			},
			options: NamingConventionOption{
				Modifiers: []string{"readonly"},
				Selector:  "parameterProperty",
			},
		},
	})
}

// cases/property.test.ts
func TestNamingConventionCasesProperty(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored { private % }`,
				`class Ignored { private "%" = 1 }`,
				`class Ignored { private readonly % = 1 }`,
				`class Ignored { private static % }`,
				`class Ignored { private static readonly % = 1 }`,
				`class Ignored { abstract % }`,
				`class Ignored { declare % }`,
				`class Ignored { #% }`,
				`class Ignored { static #% }`,
			},
			options: NamingConventionOption{
				Selector: "classProperty",
			},
		},
		{
			code: []string{
				`const ignored = { % };`,
				`const ignored = { "%": 1 };`,
			},
			options: NamingConventionOption{
				Selector: "objectLiteralProperty",
			},
		},
		{
			code: []string{
				`interface Ignored { % }`,
				`interface Ignored { "%": string }`,
				`type Ignored = { % }`,
				`type Ignored = { "%": string }`,
			},
			options: NamingConventionOption{
				Selector: "typeProperty",
			},
		},
	})
}

// cases/typeAlias.test.ts
func TestNamingConventionCasesTypeAlias(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`type % = {};`,
				`type % = 1;`,
			},
			options: NamingConventionOption{
				Selector: "typeAlias",
			},
		},
	})
}

// cases/typeParameter.test.ts
func TestNamingConventionCasesTypeParameter(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`class Ignored<%> {}`,
				`function ignored<%>() {}`,
				`type Ignored<%> = { ignored: % };`,
				`interface Ignored<%> extends Ignored<string> {}`,
			},
			options: NamingConventionOption{
				Selector: "typeParameter",
			},
		},
	})
}

// cases/variable.test.ts
func TestNamingConventionCasesVariable(t *testing.T) {
	t.Parallel()
	createTestCases(t, []testCaseSpec{
		{
			code: []string{
				`const % = 1;`,
				`let % = 1;`,
				`var % = 1;`,
				`const {%} = {ignored: 1};`,
				`const {% = 2} = {ignored: 1};`,
				`const {...%} = {ignored: 1};`,
				`const [%] = [1];`,
				`const [% = 1] = [1];`,
				`const [...%] = [1];`,
			},
			options: NamingConventionOption{
				Selector: "variable",
			},
		},
	})
}
