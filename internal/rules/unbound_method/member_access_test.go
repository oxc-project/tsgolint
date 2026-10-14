package unbound_method

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Regression cases from typescript-eslint/typescript-eslint#12448.
func TestUnboundMethodMemberAccess(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UnboundMethodRule, []rule_tester.ValidTestCase{
		{Code: `
class Foo {
  bound = () => {};
}
class Bar {
  bound = 1;
}
declare const union: Foo | Bar;
const bound = union.bound;
`},
		{Code: `
class Foo {
  bazz() {}
}
declare const foo: Foo;
declare const key: string;
const bound = foo[key];
`},
		{Code: `
class Foo {
  bazz() {}
}
declare const foo: Foo;
declare const bazz: string;
foo[bazz];
`},
		{Code: `class Foo { method(this: void) {} } declare const foo: Foo; const method = foo["method"];`},
		{Code: `class Foo { method = () => {}; } declare const foo: Foo; const method = foo["method"];`},
		{Code: `class Foo { method() {} } declare const foo: Foo; foo["method"](); const bound = foo["method"].bind(foo);`},
		{Code: `class Foo { method() {} } declare const foo: Foo; declare const key: symbol; const method = foo[key];`},
		{Code: `class Foo { 1() {} } declare const foo: Foo; declare const key: number; const method = foo[key];`},
		{Code: `class Foo { static method() {} } const method = Foo["method"];`, Options: map[string]any{"ignoreStatic": true}},
	}, []rule_tester.InvalidTestCase{
		{Code: `
class Foo {
  bazz() {}
}
class Bar {
  bazz = 1;
}
declare const union: Foo | Bar;
const bound = union.bazz;`, Errors: []rule_tester.InvalidTestCaseError{{Line: 9, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `
class Foo {
  bazz() {}
}
class Bar {
  bazz = 1;
}
declare const union: Foo | Bar;
declare const bazz: 'bazz';
const bound = union[bazz];`, Errors: []rule_tester.InvalidTestCaseError{{Line: 10, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `
class Foo {
  bazz() {}
}
declare const foo: Foo;
foo['bazz'];`, Errors: []rule_tester.InvalidTestCaseError{{Line: 6, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `
class Foo {
  bazz() {}
}
declare const foo: Foo;
declare const bazz: keyof Foo;
const bound = foo[bazz];`, Errors: []rule_tester.InvalidTestCaseError{{Line: 7, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: "\nclass Foo {\n  getValue() {}\n}\ndeclare const foo: Foo;\nconst bound = foo[`get${'Value'}`];", Errors: []rule_tester.InvalidTestCaseError{{Line: 6, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `
class Foo {
  1() {}
}
declare const foo: Foo;
foo[1];`, Errors: []rule_tester.InvalidTestCaseError{{Line: 6, MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { method() {} } declare const foo: Foo & { method: () => void }; const method = foo.method;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { method() {} } declare const foo: (Foo & { tag: true }) | { method: number }; const method = foo["method"];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { first() {} second() {} } declare const foo: Foo; declare const key: keyof Foo; const method = foo[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { bound = () => {}; method() {} } declare const foo: Foo; declare const key: keyof Foo; const method = foo[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { method(this: Foo) {} } declare const foo: Foo | { method: number }; const method = foo["method"];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unbound"}}},
		{Code: `class Foo { method = function() {}; } declare const foo: Foo | { method: number }; const method = foo.method;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unbound"}}},
		{Code: `class Foo { method() {} } declare const foo: Foo | undefined; const method = foo?.["method"];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { 1() {} } declare const foo: Foo; declare const key: 1; const method = foo[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
	})
}
