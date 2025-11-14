package prefer_includes

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferIncludesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferIncludesRule, []rule_tester.ValidTestCase{
		// Normal indexOf usage without comparison
		{Code: `
      function f(a: string): void {
        a.indexOf(b);
      }
    `},
		{Code: `
      function f(a: string): void {
        a.indexOf(b) + 0;
      }
    `},
		// Union types without includes method
		{Code: `
      function f(a: string | { value: string }): void {
        a.indexOf(b) !== -1;
      }
    `},
		// Custom type without includes method
		{Code: `
      type UserDefined = {
        indexOf(x: any): number; // don't have 'includes'
      };
      function f(a: UserDefined): void {
        a.indexOf(b) !== -1;
      }
    `},
		// Custom type with different parameters for includes
		{Code: `
      type UserDefined = {
        indexOf(x: any, fromIndex?: number): number;
        includes(x: any): boolean; // different parameters
      };
      function f(a: UserDefined): void {
        a.indexOf(b) !== -1;
      }
    `},
		{Code: `
      type UserDefined = {
        indexOf(x: any, fromIndex?: number): number;
        includes(x: any, fromIndex: number): boolean; // different parameters
      };
      function f(a: UserDefined): void {
        a.indexOf(b) !== -1;
      }
    `},
		// includes is not a method
		{Code: `
      type UserDefined = {
        indexOf(x: any, fromIndex?: number): number;
        includes: boolean; // different type
      };
      function f(a: UserDefined): void {
        a.indexOf(b) !== -1;
      }
    `},
		// Unsupported RegExp patterns
		{Code: `
      function f(a: string): void {
        /word/i.test(a);
      }
    `},
		{Code: `
      function f(a: string): void {
        /[ws]ord/.test(a);
      }
    `},
		{Code: `
      function f(a: string): void {
        /some|word/.test(a);
      }
    `},
		{Code: `
      function f(a: string): void {
        /word/.test();
      }
    `},
		{Code: `
      function f(a: string): void {
        something.test(a);
      }
    `},
		{Code: `
      const pattern = new RegExp('word');
      function f(a) {
        return pattern.test(a);
      }
    `},
	}, []rule_tester.InvalidTestCase{
		// Positive checks: !== -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// != -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) != -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// > -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) > -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// >= 0
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) >= 0;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Negative checks: === -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) === -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          !a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// == -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) == -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          !a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// <= -1
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) <= -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          !a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// < 0
		{
			Code: `
        declare const b: any;
        function f(a: string): void {
          a.indexOf(b) < 0;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: string): void {
          !a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Type variations: any[]
		{
			Code: `
        declare const b: any;
        function f(a: any[]): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: any[]): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// ReadonlyArray
		{
			Code: `
        declare const b: any;
        function f(a: ReadonlyArray<any>): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: ReadonlyArray<any>): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// TypedArrays
		{
			Code: `
        declare const b: any;
        function f(a: Int8Array): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: Int8Array): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
        declare const b: any;
        function f(a: Uint32Array): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: Uint32Array): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		{
			Code: `
        declare const b: any;
        function f(a: Float64Array): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: Float64Array): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Generic union type
		{
			Code: `
        declare const b: any;
        function f<T>(a: T[] | ReadonlyArray<T>): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f<T>(a: T[] | ReadonlyArray<T>): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Custom type with both methods
		{
			Code: `
        declare const b: any;
        type UserDefined = {
          indexOf(x: any): number;
          includes(x: any): boolean;
        };
        function f(a: UserDefined): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        type UserDefined = {
          indexOf(x: any): number;
          includes(x: any): boolean;
        };
        function f(a: UserDefined): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      8,
					Column:    11,
				},
			},
		},
		// Readonly wrapper
		{
			Code: `
        declare const b: any;
        function f(a: Readonly<any[]>): void {
          a.indexOf(b) !== -1;
        }
      `,
			Output: []string{`
        declare const b: any;
        function f(a: Readonly<any[]>): void {
          a.includes(b);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// RegExp.test() patterns - basic literal
		{
			Code: `
        function f(a: string): void {
          /word/.test(a);
        }
      `,
			Output: []string{`
        function f(a: string): void {
          a.includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
		// Test escape sequences - matching TS version test
		{
			Code: `
        function f(a: string): void {
          /\0'\\n\r\v\t\f/.test(a);
        }
      `,
			Output: []string{`
        function f(a: string): void {
          a.includes('\0\'\\n\r\v\t\f');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
		// RegExp constructor with variable reference
		{
			Code: `
        const pattern = new RegExp('word');
        function f(a: string): void {
          pattern.test(a);
        }
      `,
			Output: []string{`
        const pattern = new RegExp('word');
        function f(a: string): void {
          a.includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Binary expression argument
		{
			Code: `
        const pattern = /word/;
        function f(a: string, b: string): void {
          pattern.test(a + b);
        }
      `,
			Output: []string{`
        const pattern = /word/;
        function f(a: string, b: string): void {
          (a + b).includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Unicode characters in regex pattern
		{
			Code: `
        function f(a: string): void {
          /café/.test(a);
        }
      `,
			Output: []string{`
        function f(a: string): void {
          a.includes('café');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
		// Unicode in RegExp constructor
		{
			Code: `
        const pattern = new RegExp('café');
        function f(a: string): void {
          pattern.test(a);
        }
      `,
			Output: []string{`
        const pattern = new RegExp('café');
        function f(a: string): void {
          a.includes('café');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Test with SequenceExpression - should add parens
		{
			Code: `
        function f(a: string): void {
          /word/.test((1 + 1, a));
        }
      `,
			Output: []string{`
        function f(a: string): void {
          (1 + 1, a).includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
		// Variable reference to regex literal
		{
			Code: `
        const pattern = /word/;
        function f(a: string): void {
          pattern.test(a);
        }
      `,
			Output: []string{`
        const pattern = /word/;
        function f(a: string): void {
          a.includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Method call as argument - no parens needed
		{
			Code: `
        function getString(): string { return "test"; }
        function f(): void {
          /word/.test(getString());
        }
      `,
			Output: []string{`
        function getString(): string { return "test"; }
        function f(): void {
          getString().includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      4,
					Column:    11,
				},
			},
		},
		// Property access as argument - no parens needed
		{
			Code: `
        function f(obj: { value: string }): void {
          /word/.test(obj.value);
        }
      `,
			Output: []string{`
        function f(obj: { value: string }): void {
          obj.value.includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
		// Element access as argument - no parens needed
		{
			Code: `
        function f(arr: string[]): void {
          /word/.test(arr[0]);
        }
      `,
			Output: []string{`
        function f(arr: string[]): void {
          arr[0].includes('word');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferStringIncludes",
					Line:      3,
					Column:    11,
				},
			},
		},
	})
}
