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
