package no_deprecated

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoDeprecatedRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDeprecatedRule, []rule_tester.ValidTestCase{
		// Non-deprecated usage is valid
		{Code: `
			const value = 1;
			console.log(value);
		`},
		{Code: `
			class MyClass {
				method() {}
			}
			const instance = new MyClass();
			instance.method();
		`},
		{Code: `
			interface MyInterface {
				prop: string;
			}
			const obj: MyInterface = { prop: 'test' };
			console.log(obj.prop);
		`},
		// Declaring deprecated items is valid (the declaration itself)
		{Code: `
			/** @deprecated Use newFunction instead */
			function oldFunction() {}
		`},
		{Code: `
			class MyClass {
				/** @deprecated Use newMethod instead */
				oldMethod() {}
			}
		`},
		{Code: `
			/** @deprecated This class is deprecated */
			class OldClass {}
		`},
		{Code: `
			interface MyInterface {
				/** @deprecated */
				oldProp: string;
			}
		`},
		// Using non-deprecated overload
		{Code: `
			function fn(a: string): void;
			/** @deprecated */
			function fn(a: number): void;
			function fn(a: string | number): void {}
			
			fn('test'); // Using non-deprecated overload
		`},
	}, []rule_tester.InvalidTestCase{
		// Using deprecated variable
		{
			Code: `
				/** @deprecated */
				const oldVar = 1;
				console.log(oldVar);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    17,
				},
			},
		},
		// Using deprecated function
		{
			Code: `
				/** @deprecated Use newFunction instead */
				function oldFunction() {}
				oldFunction();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    5,
				},
			},
		},
		// Using deprecated class
		{
			Code: `
				/** @deprecated This class is old */
				class OldClass {}
				const instance = new OldClass();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      4,
					Column:    26,
				},
			},
		},
		// Using deprecated method
		{
			Code: `
				class MyClass {
					/** @deprecated Use newMethod instead */
					oldMethod() {}
					newMethod() {}
				}
				const instance = new MyClass();
				instance.oldMethod();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      8,
					Column:    14,
				},
			},
		},
		// Using deprecated property
		{
			Code: `
				class MyClass {
					/** @deprecated */
					oldProp = 1;
					newProp = 2;
				}
				const instance = new MyClass();
				console.log(instance.oldProp);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      8,
					Column:    26,
				},
			},
		},
		// Using deprecated interface property
		{
			Code: `
				interface MyInterface {
					/** @deprecated */
					oldProp: string;
					newProp: string;
				}
				const obj: MyInterface = { oldProp: 'test', newProp: 'test' };
				console.log(obj.oldProp);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      8,
					Column:    21,
				},
			},
		},
		// Using deprecated enum member
		{
			Code: `
				enum MyEnum {
					/** @deprecated */
					Old = 0,
					New = 1,
				}
				const value = MyEnum.Old;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    27,
				},
			},
		},
		// Using deprecated namespace member
		{
			Code: `
				namespace MyNamespace {
					/** @deprecated */
					export const oldValue = 1;
					export const newValue = 2;
				}
				console.log(MyNamespace.oldValue);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    30,
				},
			},
		},
		// Using deprecated type alias
		{
			Code: `
				/** @deprecated Use NewType instead */
				type OldType = string;
				type NewType = string;
				
				let value: OldType;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      6,
					Column:    16,
				},
			},
		},
		// Using deprecated with element access
		{
			Code: `
				const obj = {
					/** @deprecated */
					'old-prop': 1,
					'new-prop': 2,
				};
				console.log(obj['old-prop']);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    21,
				},
			},
		},
		// Using deprecated in destructuring
		{
			Code: `
				const obj = {
					/** @deprecated */
					oldProp: 1,
					newProp: 2,
				};
				const { oldProp } = obj;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    13,
				},
			},
		},
		// Using deprecated import
		{
			Code: `
				// file1.ts
				/** @deprecated */
				export const oldExport = 1;
				
				// file2.ts
				import { oldExport } from './file1';
				console.log(oldExport);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      8,
					Column:    17,
				},
			},
		},
		// Using deprecated with function overload
		{
			Code: `
				function fn(a: string): void;
				/** @deprecated Use string overload */
				function fn(a: number): void;
				function fn(a: string | number): void {}
				
				fn(123); // Using deprecated overload
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    5,
				},
			},
		},
		// Multiple deprecations in one line
		{
			Code: `
				/** @deprecated */
				const a = 1;
				/** @deprecated */
				const b = 2;
				console.log(a, b);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      6,
					Column:    17,
				},
				{
					MessageId: "deprecated",
					Line:      6,
					Column:    20,
				},
			},
		},
		// Deprecated static method
		{
			Code: `
				class MyClass {
					/** @deprecated */
					static oldStaticMethod() {}
					static newStaticMethod() {}
				}
				MyClass.oldStaticMethod();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      7,
					Column:    13,
				},
			},
		},
		// Deprecated constructor
		{
			Code: `
				/** @deprecated Use NewClass instead */
				class OldClass {
					constructor() {}
				}
				new OldClass();
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      6,
					Column:    9,
				},
			},
		},
		// Deprecated getter/setter
		{
			Code: `
				class MyClass {
					/** @deprecated */
					get oldValue() { return 1; }
					set oldValue(v: number) {}
					
					get newValue() { return 2; }
					set newValue(v: number) {}
				}
				const instance = new MyClass();
				console.log(instance.oldValue);
				instance.oldValue = 3;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "deprecated",
					Line:      11,
					Column:    26,
				},
				{
					MessageId: "deprecated",
					Line:      12,
					Column:    14,
				},
			},
		},
	})
}