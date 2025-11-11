package no_duplicate_type_constituents

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoDuplicateTypeConstituentsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, []rule_tester.ValidTestCase{
		{
			Code: "type T = 1 | 2;",
		},
		{
			Code: "type T = 1 | '1';",
		},
		{
			Code: "type T = true & boolean;",
		},
		{
			Code: "type T = null | undefined;",
		},
		{
			Code: "type T = any | unknown;",
		},
		{
			Code: "type T = { a: string } | { b: string };",
		},
		{
			Code: "type T = { a: string; b: number } | { b: number; a: string };",
		},
		{
			Code: "type T = { a: string | number };",
		},
		{
			Code: "type T = Set<string> | Set<number>;",
		},
		{
			Code: "type T = Class<string> | Class<number>;",
		},
		{
			Code: "type T = string[] | number[];",
		},
		{
			Code: "type T = string[][] | string[];",
		},
		{
			Code: "type T = [1, 2, 3] | [1, 2, 4];",
		},
		{
			Code: "type T = [1, 2, 3] | [1, 2, 3, 4];",
		},
		{
			Code: "type T = 'A' | string[];",
		},
		{
			Code: "type T = (() => string) | (() => void);",
		},
		{
			Code: "type T = () => string | void;",
		},
		{
			Code: "type T = () => null | undefined;",
		},
		{
			Code: "type T = (arg: string | number) => void;",
		},
		{
			Code: "type T = A | A;",
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B;
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
const a: A | B = 'A';
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | /* comment */ B;
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = 'A' | 'B';
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | C;
      `,
		},
		{
			Code: "type T = readonly string[] | string[];",
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type T = (A | B) | (C | D);
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A & B);
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = Record<string, A | B>;
      `,
		},
		{
			Code:    "type T = A | A;",
			Options: rule_tester.OptionsFromJSON[NoDuplicateTypeConstituentsOptions](`{"ignoreUnions": true}`),
		},
		{
			Code:    "type T = A & A;",
			Options: rule_tester.OptionsFromJSON[NoDuplicateTypeConstituentsOptions](`{"ignoreIntersections": true}`),
		},
		{
			Code: "type T = Class<string> | Class<string>;",
		},
		{
			Code: "type T = A | A | string;",
		},
		{
			Code: "(a: string | undefined) => {};",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "type T = 1 | 1;",
			Output: []string{"type T = 1  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = true & true;",
			Output: []string{"type T = true  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = null | null;",
			Output: []string{"type T = null  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = any | any;",
			Output: []string{"type T = any  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string | string };",
			Output: []string{"type T = { a: string   };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string } | { a: string };",
			Output: []string{"type T = { a: string }  ;"},
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string; b: number } | { a: string; b: number };",
			Output: []string{"type T = { a: string; b: number }  ;"},
			Skip:   true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = Set<string> | Set<string>;",
			Output: []string{"type T = Set<string>  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number> | IsArray<string>;
      `,
			Output: []string{`
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number>  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = string[] | string[];",
			Output: []string{"type T = string[]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = string[][] | string[][];",
			Output: []string{"type T = string[][]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = [1, 2, 3] | [1, 2, 3];",
			Output: []string{"type T = [1, 2, 3]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = () => string | string;",
			Output: []string{"type T = () => string  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = () => null | null;",
			Output: []string{"type T = () => null  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = (arg: string | string) => void;",
			Output: []string{"type T = (arg: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = 'A' | 'A';",
			Output: []string{"type T = 'A'  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | A;
      `,
			Output: []string{`
type A = 'A';
type T = A  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
const a: A | A = 'A';
      `,
			Output: []string{`
type A = 'A';
const a: A   = 'A';
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | /* comment */ A;
      `,
			Output: []string{`
type A = 'A';
type T = A  /* comment */ ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1 | A2 | A3;
      `,
			Output: []string{`
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A | B;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A | A;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | A | C;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B   | C;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = (A | B)  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | (A | A);
      `,
			Output: []string{`
type A = 'A';
type T = A  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B) | (A | B) | ((C | D) & (A | B)) | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B)   | ((C | D) & (A | B))  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = (A | B) | A | B | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = (A | B)      ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number | string) | number | string;",
			Output: []string{"type A = (number | string)    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number | (string | null)) | (string | (null | number));",
			Output: []string{"type A = (number | (string | null))  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number & string) & number & string;",
			Output: []string{"type A = (number & string)    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: "type A = number & string & (number & string);",
			Output: []string{"type A = number & string & (  string);",
				"type A = number & string  ;",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = Record<string, A | A>;
      `,
			Output: []string{`
type A = 'A';
type T = Record<string, A  >;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = A | A | string | string;",
			Output: []string{"type T = A | A | string  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "(a?: string | undefined) => {};",
			Output: []string{"(a?: string  ) => {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        type T = undefined;
        (arg?: T | string) => {};
      `,
			Output: []string{`
        type T = undefined;
        (arg?:   string) => {};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface F {
          (a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface F {
          (a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = new (a?: string | undefined) => void;",
			Output: []string{"type fn = new (a?: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "function f(a?: string | undefined) {}",
			Output: []string{"function f(a?: string  ) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "f = function (a?: string | undefined) {};",
			Output: []string{"f = function (a?: string  ) {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "declare function f(a?: string | undefined): void;",
			Output: []string{"declare function f(a?: string  ): void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        declare class bb {
          f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        declare class bb {
          f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface ee {
          f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface ee {
          f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface ee {
          new (a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface ee {
          new (a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: string | undefined) => void;",
			Output: []string{"type fn = (a?: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: string | (undefined | number)) => void;",
			Output: []string{"type fn = (a?: string | (  number)) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: (undefined | number) | string) => void;",
			Output: []string{"type fn = (a?: (  number) | string) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        abstract class cc {
          abstract f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        abstract class cc {
          abstract f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
	})
}
