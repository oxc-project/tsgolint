package require_using_for_disposable

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestRequireUsingForDisposableRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RequireUsingForDisposableRule, []rule_tester.ValidTestCase{
		// Using keyword present for disposable function
		{Code: `
        function disposable() {
          return {
            [Symbol.dispose]: () => {
              console.log('dispose');
            }
          };
        }

        using result = disposable();
      `},
		// Using keyword present for async disposable function
		{Code: `
        function disposable() {
          return {
            [Symbol.asyncDispose]: async () => {
              console.log('dispose');
            }
          };
        }

        await using result = disposable();
      `},
	}, []rule_tester.InvalidTestCase{
		// Missing using keyword in disposable function
		{
			Code: `
        function disposable() {
          return {
            [Symbol.dispose]: () => {
              console.log('dispose');
            }
          };
        }

        const result = disposable();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireUsing",
				},
			},
		},
		// Missing using keyword in async disposable function
		{
			Code: `
        function disposable() {
          return {
            [Symbol.asyncDispose]: async () => {
              console.log('dispose');
            }
          };
        }

        const result = disposable();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireUsing",
				},
			},
		},
		// Missing using keyword in disposable class
		{
			Code: `
        class DisposableClass implements Disposable {
          [Symbol.dispose]() {
            console.log('Cleanup');
          }
        }

        const cls = new DisposableClass();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireUsing",
				},
			},
		},
		// Missing using keyword in async disposable class
		{
			Code: `
        class DisposableClass implements AsyncDisposable {
          async [Symbol.asyncDispose]() {
            console.log('Cleanup');
          }
        }

        const cls = new DisposableClass();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireUsing",
				},
			},
		},
		// Missing using keyword for union type
		{
			Code: `
        function disposable(skipDispose: boolean) {
          if (skipDispose) {
            return 10;
          }

          return {
            [Symbol.dispose]: () => {
              console.log('dispose');
            },
          };
        }

        const result = disposable(false);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "requireUsing",
				},
			},
		},
	})
}
