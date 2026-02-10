package no_invalid_void_type

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoInvalidVoidTypeRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInvalidVoidTypeRule, []rule_tester.ValidTestCase{
		{Code: "type Generic<T> = [T];", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`)},
		{Code: "type voidNeverUnion = void | never;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`)},
		{Code: "type neverVoidUnion = never | void;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`)},
		{Code: "function func(): void {}"},
		{Code: "type Fn = () => void;"},
		{Code: "let voidPromise: Promise<void> = new Promise<void>(() => {});"},
		{Code: "let voidMap: Map<string, void> = new Map<string, void>();"},
		{Code: "type Generic<T> = [T]; type GenericVoid = Generic<void>;"},
		{Code: "const arrowGeneric1 = <T = void,>(arg: T) => {};"},
		{Code: "declare function functionDeclaration1<T = void>(arg: T): void;"},
		{Code: "type voidPromiseUnion = void | Promise<void>;"},
		{Code: "type promiseNeverUnion = Promise<void> | never;"},
		{Code: "type Generic<T> = [T]; type GenericVoid = Generic<void>;"},
		{Code: "type Allowed<T> = [T]; type AllowedVoid = Allowed<void>;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Allowed"]}`)},
		{Code: "type AllowedVoid = Ex.Mx.Tx<void>;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Ex.Mx.Tx"]}`)},
		{Code: "type AllowedVoid = Ex . Mx . Tx<void>;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Ex . Mx . Tx"]}`)},
		{Code: "type AllowedVoidUnion = void | Ex.Mx.Tx<void>;", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Ex.Mx.Tx"]}`)},
		{Code: "function f(this: void) {}", Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowAsThisParameter": true}`)},
		{Code: `
class Test {
  public static helper(this: void) {}
  method(this: void) {}
}
      `, Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowAsThisParameter": true}`)},
		{Code: `
function f(): void;
function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) {
    return x;
  }
}
      `},
		{Code: `
export default function (): void;
export default function (x: string): string;
export default function (x?: string): string | void {
  if (x !== undefined) {
    return x;
  }
}
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "type GenericVoid = Generic<void>;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturn"},
			},
		},
		{
			Code:    "function takeVoid(thing: void) {}",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturn"},
			},
		},
		{
			Code:    "type invalidVoidUnion = void | number;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturn"},
			},
		},
		{
			Code: "function takeVoid(thing: void) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrGeneric"},
			},
		},
		{
			Code: "functionGeneric<void>(undefined);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrGeneric"},
			},
		},
		{
			Code: "type InvalidVoidUnion = string | void;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidUnionConstituent"},
			},
		},
		{
			Code: "const arrowGeneric = <T extends void>(arg: T) => {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrGeneric"},
			},
		},
		{
			Code: "declare function test<T extends number | void>(): T;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidUnionConstituent"},
			},
		},
		{
			Code:    "type Banned<T> = [T]; type BannedVoid = Banned<void>;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Allowed"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidForGeneric"},
			},
		},
		{
			Code:    "type BannedVoid = Ex.Mx.Tx<void>;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Tx"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidForGeneric"},
			},
		},
		{
			Code:    "type BannedUnion = void | Promise<void>;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowInGenericTypeArguments": ["Allowed"]}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidUnionConstituent"},
				{MessageId: "invalidVoidForGeneric"},
			},
		},
		{
			Code: "function f(this: void) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrGeneric"},
			},
		},
		{
			Code:    "type Alias = void;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowAsThisParameter": true, "allowInGenericTypeArguments": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrThisParam"},
			},
		},
		{
			Code:    "type alias = void;",
			Options: rule_tester.OptionsFromJSON[NoInvalidVoidTypeOptions](`{"allowAsThisParameter": true, "allowInGenericTypeArguments": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidNotReturnOrThisParamOrGeneric"},
			},
		},
		{
			Code: `
export default function (x?: string): string | void {
  if (x !== undefined) {
    return x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidVoidUnionConstituent"},
			},
		},
	})
}
