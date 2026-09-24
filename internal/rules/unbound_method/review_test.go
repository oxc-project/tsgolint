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

func TestUnboundMethodComputedNativeBound(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &UnboundMethodRule, []rule_tester.ValidTestCase{
		{Code: `const collator = new Intl.Collator(); const compare = collator['compare'];`},
		{Code: `const collator = new Intl.Collator(); declare const key: 'compare'; const compare = collator[key];`},
		{Code: `class Derived extends Intl.Collator {} const compare = new Derived()['compare'];`},
		{Code: `const floor = Math['floor']; const math = Math; const ceil = math['ceil'];`},
		{Code: `declare const collator: Intl.Collator | { compare: number }; const compare = collator['compare'];`},
	}, []rule_tester.InvalidTestCase{
		{Code: `const collator = new Intl.Collator(); const method = collator['resolvedOptions'];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `class Derived extends Intl.Collator { compare(a: string, b: string) { return 0; } } const compare = new Derived()['compare'];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `interface Custom { compare(a: string, b: string): number; } declare const collator: Intl.Collator | Custom; const compare = collator['compare'];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `const collator = new Intl.Collator(); declare const key: 'compare' | 'resolvedOptions'; const method = collator[key];`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
		{Code: `const Math = { floor() {} }; const floor = Math['floor']; export {};`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation"}}},
	})
}
