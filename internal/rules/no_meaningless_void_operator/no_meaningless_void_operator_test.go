package no_meaningless_void_operator

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoMeaninglessVoidOperatorRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoMeaninglessVoidOperatorRule, []rule_tester.ValidTestCase{
		{Code: `
(() => {})();

function foo() {}
foo(); // nothing to discard

function bar() {
  return 2;
}
void bar(); // discarding a number
    `},
		{Code: `declare function fail(): never; void fail();`},
		{Code: `void 0; void (0x0 as number);`},
		{Code: `declare const promise: Promise<void>; void promise; void new Promise<void>(() => {});`},
		{Code: `declare const thenable: { then(onFulfilled: () => void): void }; void thenable;`},
		{Code: `void import('./file'); export {};`},
		{Code: `declare function getValue(): string; void (getValue() as string);`},
		// Assignment regressions from typescript-eslint/typescript-eslint#12873.
		{Code: `declare let x: number; void (x = 1);`},
		{Code: `declare let x: number; () => void (x = 1);`},
		{Code: `declare let x: number; void (x += 1);`},
		{Code: `declare let x: number; declare let y: number; void (x = y = 1);`},
		{Code: `declare let x: string; declare function getValue(): string; void (x = getValue());`},
		{Code: `declare const obj: { prop: number }; void (obj.prop = 1);`},
		{Code: `declare let x: number; declare let y: number; void ((x = 1), (y = 2));`},
		{Code: `declare let x: undefined; () => void (((x = undefined)));`},
		{Code: `declare let x: void; declare function fn(): void; void (x = fn());`},
		{Code: `declare let x: undefined; void (x ??= undefined);`},
		{Code: `declare let x: undefined; void ((x = undefined) as undefined);`},
		{Code: `declare let x: undefined; void (<undefined>(x = undefined));`},
		{Code: `declare let x: undefined; void ((x = undefined) satisfies undefined);`},
		{Code: `declare let x: undefined; void ((x = undefined), (x = undefined));`},
		{Code: `declare let x: undefined; void ((x = undefined)!, ((x = undefined) as undefined));`},
		{
			Code:    `declare let x: undefined; void ((x = undefined)!);`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
		},
		{
			Code:    `declare let x: never; declare function fail(): never; void (x = fail());`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "void (() => {})();",
			Output: []string{" (() => {})();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: `
function foo() {}
void foo();
      `,
			Output: []string{`
function foo() {}
 foo();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code: `declare function fail(): never;
void fail();`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "removeVoid",
							Output: `declare function fail(): never;
 fail();`,
						},
					},
				},
			},
		},
		{
			Code: `
const foo = (() => {}) as (() => void) | undefined;
void foo?.();
      `,
			Output: []string{`
const foo = (() => {}) as (() => void) | undefined;
 foo?.();
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "meaninglessVoidOperator",
					Line:      3,
					Column:    1,
				},
			},
		},
		{
			Code:   `declare let x: undefined; declare function fn(): void; void ((x = undefined), fn());`,
			Output: []string{`declare let x: undefined; declare function fn(): void;  ((x = undefined), fn());`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOperator"}},
		},
		{
			Code:   `function handler(): void {} void handler; export {};`,
			Output: []string{`function handler(): void {}  handler; export {};`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `function bar(x: number) { void x; }`,
			Output: []string{`function bar(x: number) {  x; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:    `function bar(x: never) { void x; }`,
			Options: rule_tester.OptionsFromJSON[NoMeaninglessVoidOperatorOptions](`{"checkNever": true}`),
			Output:  []string{`function bar(x: never) {  x; }`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `declare const obj: { value: number } | undefined; (void (0, obj?.value)!);`,
			Output: []string{`declare const obj: { value: number } | undefined; ( (0, obj?.value)!);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `const result = void 1;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `const result = void (import('./file') as unknown); export {};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `void function() {};`,
			Output: []string{`( function() {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `void { a: 1, b: 2 };`,
			Output: []string{`( { a: 1, b: 2 });`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `if (true) void /* keep */ async function() {};`,
			Output: []string{`if (true) ( /* keep */ async function() {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `void class {}.name;`,
			Output: []string{`( class {}.name);`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code:   `void "use strict";`,
			Output: []string{`( "use strict");`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "meaninglessVoidOnNonCall"}},
		},
		{
			Code: `run()
void { a: 1 };
run()
void [1, 2];`,
			Output: []string{`run()
;( { a: 1 });
run()
; [1, 2];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "meaninglessVoidOnNonCall"},
				{MessageId: "meaninglessVoidOnNonCall"},
			},
		},
	})
}
