package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestUpstreamMoreVariations tests additional systematic variations
// Focus: nested structures, multiple operators, complex type scenarios
func TestUpstreamMoreVariations(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// === Await/yield patterns where base is await expression itself ===
		// These are not safe to convert because await at the base
		{Code: `await foo && (await foo).bar;`},
		{Code: `(await foo) && (await foo).bar;`},
		{Code: `(yield foo) && (yield foo).bar;`},
	},
		[]rule_tester.InvalidTestCase{
			// === Nested object access patterns ===

			{
				Code:    `a.b && a.b.c;`,
				Output:  []string{`a.b?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a.b.c && a.b.c.d;`,
				Output:  []string{`a.b.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a.b && a.b.c && a.b.c.d;`,
				Output:  []string{`a.b?.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a.b.c && a.b.c.d && a.b.c.d.e;`,
				Output:  []string{`a.b.c?.d?.e;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// With nullish checks
			{
				Code:    `a.b != null && a.b.c;`,
				Output:  []string{`a.b?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a.b.c != null && a.b.c.d;`,
				Output:  []string{`a.b.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `a.b != null && a.b.c != null && a.b.c.d;`,
				Output:  []string{`a.b?.c?.d;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Array-like access patterns ===

			{
				Code:    `arr[0] && arr[0].prop;`,
				Output:  []string{`arr[0]?.prop;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `arr[0] && arr[0][1];`,
				Output:  []string{`arr[0]?.[1];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `arr[0] && arr[0][1] && arr[0][1].prop;`,
				Output:  []string{`arr[0]?.[1]?.prop;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `matrix[i] && matrix[i][j];`,
				Output:  []string{`matrix[i]?.[j];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `matrix[i] && matrix[i][j] && matrix[i][j][k];`,
				Output:  []string{`matrix[i]?.[j]?.[k];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Method chaining patterns ===

			{
				Code:    `obj.method() && obj.method().prop;`,
				Output:  []string{`obj.method()?.prop;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.method1() && obj.method1().method2();`,
				Output:  []string{`obj.method1()?.method2();`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.method1() && obj.method1().method2() && obj.method1().method2().prop;`,
				Output:  []string{`obj.method1()?.method2()?.prop;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.a.b() && obj.a.b().c;`,
				Output:  []string{`obj.a.b()?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.a && obj.a.b() && obj.a.b().c;`,
				Output:  []string{`obj.a?.b()?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Template literal / computed property patterns ===

			{
				Code:    "obj && obj[`key`];",
				Output:  []string{"obj?.[`key`];"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "obj && obj[`key`] && obj[`key`].prop;",
				Output:  []string{"obj?.[`key`]?.prop;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "obj && obj[`prefix_${id}`];",
				Output:  []string{"obj?.[`prefix_${id}`];"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    "obj && obj[`prefix_${id}`] && obj[`prefix_${id}`].value;",
				Output:  []string{"obj?.[`prefix_${id}`]?.value;"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Comparison operator at end variations ===

			{
				Code:    `foo && foo.bar === value;`,
				Output:  []string{`foo?.bar === value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar !== value;`,
				Output:  []string{`foo?.bar !== value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar == value;`,
				Output:  []string{`foo?.bar == value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar != value;`,
				Output:  []string{`foo?.bar != value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar > value;`,
				Output:  []string{`foo?.bar > value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar < value;`,
				Output:  []string{`foo?.bar < value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar >= value;`,
				Output:  []string{`foo?.bar >= value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo && foo.bar <= value;`,
				Output:  []string{`foo?.bar <= value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// Multi-level with comparison at end
			{
				Code:    `foo && foo.bar && foo.bar.baz === value;`,
				Output:  []string{`foo?.bar?.baz === value;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar.baz && foo.bar.baz.qux > 0;`,
				Output:  []string{`foo.bar?.baz?.qux > 0;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === instanceof checks ===

			{
				Code:    `foo && foo.bar instanceof SomeClass;`,
				Output:  []string{`foo?.bar instanceof SomeClass;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo.bar && foo.bar.baz instanceof Error;`,
				Output:  []string{`foo.bar?.baz instanceof Error;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Ternary operator patterns ===

			{
				Code:    `foo && foo.bar ? foo.bar.baz : defaultValue;`,
				Output:  []string{`foo?.bar ? foo.bar.baz : defaultValue;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Assignment patterns ===

			{
				Code:    `let x = foo && foo.bar;`,
				Output:  []string{`let x = foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `const y = foo && foo.bar && foo.bar.baz;`,
				Output:  []string{`const y = foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `var z = foo.bar && foo.bar.baz;`,
				Output:  []string{`var z = foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Return statements ===

			{
				Code:    `return foo && foo.bar;`,
				Output:  []string{`return foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `return foo && foo.bar && foo.bar.baz;`,
				Output:  []string{`return foo?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `return foo.bar && foo.bar.baz;`,
				Output:  []string{`return foo.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Function arguments ===

			{
				Code:    `doSomething(foo && foo.bar);`,
				Output:  []string{`doSomething(foo?.bar);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `doSomething(foo && foo.bar, bar && bar.baz);`,
				Output: []string{`doSomething(foo?.bar, bar?.baz);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `func(foo && foo.bar && foo.bar.baz);`,
				Output:  []string{`func(foo?.bar?.baz);`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Nested in arrays ===

			{
				Code:    `[foo && foo.bar];`,
				Output:  []string{`[foo?.bar];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `[foo && foo.bar, bar && bar.baz];`,
				Output: []string{`[foo?.bar, bar?.baz];`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `[foo && foo.bar && foo.bar.baz];`,
				Output:  []string{`[foo?.bar?.baz];`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Nested in objects ===

			{
				Code:    `({key: foo && foo.bar});`,
				Output:  []string{`({key: foo?.bar});`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:   `({a: foo && foo.bar, b: bar && bar.baz});`,
				Output: []string{`({a: foo?.bar, b: bar?.baz});`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferOptionalChain"},
					{MessageId: "preferOptionalChain"},
				},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `({prop: foo && foo.bar && foo.bar.baz});`,
				Output:  []string{`({prop: foo?.bar?.baz});`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === With await - property access on await result IS safe ===
			// Note: (await foo) && (await foo).bar is NOT safe (base is await expression)
			// But (await foo).bar && (await foo).bar.baz IS safe (base is property access)
			{
				Code:   `(await foo) && (await foo).bar && (await foo).bar.baz;`,
				Output: []string{`(await foo) && (await foo).bar?.baz;`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{
					"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
				},
			},

			// === With new ===

			{
				Code:    `new Foo() && new Foo().bar;`,
				Output:  []string{`new Foo()?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `new Foo() && new Foo().bar && new Foo().bar.baz;`,
				Output:  []string{`new Foo()?.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === TypeScript as/satisfies expressions ===

			{
				Code:    `(foo as any) && (foo as any).bar;`,
				Output:  []string{`(foo as any)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `(foo as SomeType) && (foo as SomeType).bar;`,
				Output:  []string{`(foo as SomeType)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `(<SomeType>foo) && (<SomeType>foo).bar;`,
				Output:  []string{`(<SomeType>foo)?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Non-null assertion patterns ===

			{
				Code:    `foo! && foo!.bar;`,
				Output:  []string{`foo!?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo!.bar && foo!.bar.baz;`,
				Output:  []string{`foo!.bar?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `foo! && foo!.bar! && foo!.bar!.baz;`,
				Output:  []string{`foo!?.bar!?.baz;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Mixed contexts ===

			{
				Code: `
if (foo && foo.bar) {
  console.log(foo.bar);
}`,
				Output: []string{`
if (foo?.bar) {
  console.log(foo.bar);
}`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
while (foo && foo.bar) {
  doSomething();
}`,
				Output: []string{`
while (foo?.bar) {
  doSomething();
}`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code: `
for (let i = 0; foo && foo.bar; i++) {
  doSomething();
}`,
				Output: []string{`
for (let i = 0; foo?.bar; i++) {
  doSomething();
}`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Switch cases ===

			{
				Code: `
switch (foo && foo.bar) {
  case value: break;
}`,
				Output: []string{`
switch (foo?.bar) {
  case value: break;
}`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Try-catch ===

			{
				Code: `
try {
  result = foo && foo.bar;
} catch (e) {}`,
				Output: []string{`
try {
  result = foo?.bar;
} catch (e) {}`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Throw statements ===

			{
				Code:    `throw foo && foo.bar;`,
				Output:  []string{`throw foo?.bar;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},

			// === Deep nesting ===

			{
				Code:    `a && a.b && a.b.c && a.b.c.d && a.b.c.d.e && a.b.c.d.e.f;`,
				Output:  []string{`a?.b?.c?.d?.e?.f;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.a && obj.a.b && obj.a.b.c && obj.a.b.c.d && obj.a.b.c.d.e && obj.a.b.c.d.e.f;`,
				Output:  []string{`obj.a?.b?.c?.d?.e?.f;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
			{
				Code:    `obj.x.y && obj.x.y.z && obj.x.y.z.a && obj.x.y.z.a.b && obj.x.y.z.a.b.c;`,
				Output:  []string{`obj.x.y?.z?.a?.b?.c;`},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
				Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
			},
		})
}
