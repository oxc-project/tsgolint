package no_unsafe_unary_minus

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeUnaryMinusRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeUnaryMinusRule, []rule_tester.ValidTestCase{
		{Code: "+42;"},
		{Code: "-42;"},
		{Code: "-42n;"},
		{Code: "(a: number) => -a;"},
		{Code: "(a: bigint) => -a;"},
		{Code: "(a: number | bigint) => -a;"},
		{Code: "(a: any) => -a;"},
		{Code: "(a: 1 | 2) => -a;"},
		{Code: "(a: string) => +a;"},
		{Code: "(a: number[]) => -a[0];"},
		{Code: "<T,>(t: T & number) => -t;"},
		{Code: "(a: { x: number }) => -a.x;"},
		{Code: "(a: never) => -a;"},
		{Code: "<T extends number>(t: T) => -t;"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "(a: string) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "(a: {}) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "(a: number[]) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "-'hello';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "-`hello`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "(a: { x: number }) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "(a: unknown) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "(a: void) => -a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
		{
			Code: "<T,>(t: T) => -t;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unaryMinus",
				},
			},
		},
	})
}
