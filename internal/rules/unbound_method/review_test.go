package unbound_method

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestUnboundMethodPrivateAccess(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UnboundMethodRule, []rule_tester.ValidTestCase{
		{Code: `class Foo { #method() {} run() { this.#method(); } }`},
		{Code: `class Foo { #method(this: void) {} run() { return this.#method; } }`},
		{Code: `class Foo { #method = () => {}; run() { return this.#method; } }`},
	}, []rule_tester.InvalidTestCase{
		{Code: `class Foo { #method() {} run() { return this.#method; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Foo { #method(this: Foo) {} run() { return this.#method; } }`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unbound"}}},
	})
}

func TestUnboundMethodSymbolAccess(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UnboundMethodRule, []rule_tester.ValidTestCase{
		{Code: `const key = Symbol(); class Foo { [key]() {} } declare const foo: Foo; foo[key]();`},
		{Code: `const key = Symbol(); class Foo { [key] = () => {}; } declare const foo: Foo; const method = foo[key];`},
		{Code: `const key = Symbol(); class Foo { [key](this: void) {} } declare const foo: Foo; const method = foo[key];`},
		{Code: `class Foo { method() {} } declare const foo: Foo; declare const key: symbol; const method = foo[key];`},
	}, []rule_tester.InvalidTestCase{
		{Code: `const key = Symbol(); class Foo { [key]() {} } declare const foo: Foo; const method = foo[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `const key = Symbol(); class Foo { [key]() {} } declare const foo: Foo | { [key]: number }; const method = foo[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `const key = Symbol(); class Foo { [key]() {} other() {} } declare const foo: Foo; declare const name: keyof Foo; const method = foo[name];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
	})
}
