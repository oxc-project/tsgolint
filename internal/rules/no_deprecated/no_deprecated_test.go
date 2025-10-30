package no_deprecated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
)

func TestNoDeprecated(t *testing.T) {
	rule_tester.Run(
		t,
		NoDeprecatedRule,
		map[string]rule_tester.ValidTestCase{
			// Declaring deprecated items should be allowed
			"deprecated variable declaration": {
				Code: `/** @deprecated */ var a = 1;`,
			},
			"deprecated const declaration": {
				Code: `/** @deprecated */ const a = 1;`,
			},
			"deprecated function declaration": {
				Code: `/** @deprecated */ function foo() {}`,
			},
			"deprecated class declaration": {
				Code: `/** @deprecated */ class A {}`,
			},

			// Using non-deprecated items should be allowed
			"non-deprecated variable": {
				Code: `
					const a = 1;
					const b = a;
				`,
			},
			"non-deprecated function": {
				Code: `
					function foo() { return 1; }
					foo();
				`,
			},
			"non-deprecated method": {
				Code: `
					class A {
						method() { return 1; }
					}
					new A().method();
				`,
			},
			"non-deprecated property": {
				Code: `
					const obj = { prop: 1 };
					obj.prop;
				`,
			},

			// Importing deprecated items (imports themselves shouldn't error)
			"importing deprecated": {
				Code: `
					declare module 'deprecations' {
						/** @deprecated */
						export const value = true;
					}
					import { value } from 'deprecations';
				`,
			},
		},
		map[string]rule_tester.InvalidTestCase{
			// Using deprecated variable
			"deprecated variable usage": {
				Code: `
					/** @deprecated */
					var a = undefined;
					a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Using deprecated const
			"deprecated const usage": {
				Code: `
					/** @deprecated */
					const a = { b: 1 };
					const c = a;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    16,
					},
				},
			},

			// Using deprecated function
			"deprecated function call": {
				Code: `
					/** @deprecated */
					function foo() { return 1; }
					foo();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Using deprecated with reason
			"deprecated with reason": {
				Code: `
					/** @deprecated Use newFunc instead. */
					function oldFunc() { return 1; }
					oldFunc();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecatedWithReason",
						Line:      4,
						Column:    6,
					},
				},
			},

			// Using deprecated class constructor
			"deprecated class": {
				Code: `
					/** @deprecated */
					class A {}
					new A();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      4,
						Column:    10,
					},
				},
			},

			// Using deprecated property
			"deprecated property access": {
				Code: `
					const a = {
						/** @deprecated */
						b: { c: 1 },
					};
					a.b.c;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      6,
						Column:    8,
					},
				},
			},

			// Using deprecated method
			"deprecated method call": {
				Code: `
					declare class A {
						/** @deprecated */
						method(): string;
					}
					declare const a: A;
					a.method();
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},

			// Using deprecated enum member
			"deprecated enum member": {
				Code: `
					enum A {
						/** @deprecated */
						Old,
						New
					}
					A.Old;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},

			// Using deprecated namespace member
			"deprecated namespace member": {
				Code: `
					namespace A {
						/** @deprecated */
						export const old = '';
						export const current = '';
					}
					A.old;
				`,
				Errors: []rule_tester.ExpectedError{
					{
						MessageId: "deprecated",
						Line:      7,
						Column:    8,
					},
				},
			},
		},
	)
}
