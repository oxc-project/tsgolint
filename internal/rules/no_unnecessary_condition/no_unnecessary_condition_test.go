package no_unnecessary_condition

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestNoUnnecessaryConditionRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryConditionRule, []rule_tester.ValidTestCase{
		// Variables with proper nullable types
		{Code: `
      declare const foo: string | undefined;
      if (foo) {}
    `},
		{Code: `
      declare const foo: number | null;
      if (foo) {}
    `},
		{Code: `
      declare const foo: boolean | undefined;
      if (foo) {}
    `},
		{Code: `
      declare const foo: object | null;
      if (foo) {}
    `},

		// Proper optional chaining
		{Code: `
      declare const foo: { bar?: string };
      foo.bar?.trim();
    `},
		{Code: `
      declare const foo: Array<string> | undefined;
      foo?.[0];
    `},
		{Code: `
      declare const foo: (() => void) | undefined;
      foo?.();
    `},

		// Boolean expressions with unions
		{Code: `
      declare const foo: string | number;
      if (foo) {}
    `},
		{Code: `
      declare const foo: boolean;
      if (foo) {}
    `},

		// Logical expressions
		{Code: `
      declare const foo: string | undefined;
      const bar = foo || 'default';
    `},
		{Code: `
      declare const foo: number | null;
      const bar = foo && foo > 0;
    `},

		// Loop conditions with allowed constant literals
		{
			Code: `
        while (true) {
          break;
        }
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref("only-allowed-literals")},
		},
		{
			Code: `
        for (; false; ) {}
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref("only-allowed-literals")},
		},
		{
			Code: `
        do {} while (0);
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref("only-allowed-literals")},
		},
		{
			Code: `
        while (1) {
          break;
        }
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref("only-allowed-literals")},
		},

		// Constant loop conditions with always option
		{
			Code: `
        while (true) {
          break;
        }
      `,
			Options: NoUnnecessaryConditionOptions{AllowConstantLoopConditions: utils.Ref("always")},
		},

		// Type parameters with proper constraints
		{Code: `
      function test<T>(arg: T) {
        if (arg) {}
      }
    `},
		{Code: `
      function test<T extends string | undefined>(arg: T) {
        if (arg) {}
      }
    `},

		// Any type
		{Code: `
      declare const foo: any;
      if (foo) {}
    `},

		// Negation operator
		{Code: `
      declare const foo: string | undefined;
      if (!foo) {}
    `},

		// Conditional expressions
		{Code: `
      declare const foo: number | null;
      const bar = foo ? foo + 1 : 0;
    `},
	}, []rule_tester.InvalidTestCase{
		// Always truthy - objects
		{
			Code: `
        declare const foo: object;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},
		{
			Code: `
        declare const foo: { bar: string };
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},
		{
			Code: `
        declare const foo: Array<string>;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Always truthy - non-empty string literals
		{
			Code: `
        declare const foo: 'hello';
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Always truthy - non-zero number literals
		{
			Code: `
        declare const foo: 42;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Always truthy - true literal
		{
			Code: `
        declare const foo: true;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Always falsy - null/undefined
		{
			Code: `
        declare const foo: null;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},
		{
			Code: `
        declare const foo: undefined;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// Always falsy - false literal
		{
			Code: `
        declare const foo: false;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// Always falsy - zero
		{
			Code: `
        declare const foo: 0;
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// Always falsy - empty string
		{
			Code: `
        declare const foo: '';
        if (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// Unnecessary optional chain on non-nullable
		{
			Code: `
        declare const foo: { bar: string };
        foo.bar?.trim();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "neverOptionalChain",
				},
			},
		},
		{
			Code: `
        declare const foo: Array<string>;
        foo?.[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "neverOptionalChain",
				},
			},
		},
		{
			Code: `
        declare const foo: () => void;
        foo?.();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "neverOptionalChain",
				},
			},
		},

		// While loop with always truthy
		{
			Code: `
        declare const foo: object;
        while (foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Do-while loop with always falsy
		{
			Code: `
        declare const foo: null;
        do {} while (foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// For loop with always truthy
		{
			Code: `
        declare const foo: true;
        for (; foo; ) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Conditional expression with always truthy
		{
			Code: `
        declare const foo: object;
        const bar = foo ? 1 : 2;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Logical AND with always truthy
		{
			Code: `
        declare const foo: object;
        const bar = foo && true;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Logical OR with always falsy
		{
			Code: `
        declare const foo: null;
        const bar = foo || 'default';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysFalsy",
				},
			},
		},

		// Negation with always truthy
		{
			Code: `
        declare const foo: object;
        if (!foo) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Type parameter with constraint
		{
			Code: `
        function test<T extends object>(arg: T) {
          if (arg) {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// Constant loop conditions - not allowed by default
		{
			Code: `
        while (true) {
          break;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alwaysTruthy",
				},
			},
		},

		// No strictNullChecks
		{
			Code: `
        function foo(): boolean {}
      `,
			Options:  NoUnnecessaryConditionOptions{AllowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: utils.Ref(false)},
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
				},
			},
		},
	})
}
