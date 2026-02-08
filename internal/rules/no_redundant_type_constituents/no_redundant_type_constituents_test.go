package no_redundant_type_constituents

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoRedundantTypeConstituentsRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoRedundantTypeConstituentsRule, []rule_tester.ValidTestCase{
		{Code: `
      type T = any;
      type U = T;
    `},
		{Code: `
      type T = never;
      type U = T;
    `},
		{Code: `
      type T = 1 | 2;
      type U = T | 3;
      type V = U;
    `},
		{Code: "type T = () => never;"},
		{Code: "type T = () => never | string;"},
		{Code: `
      type B = never;
      type T = () => B | string;
    `},
		{Code: `
      type B = string;
      type T = () => B | never;
    `},
		{Code: "type T = () => string | never;"},
		{Code: "type T = { (): string | never };"},
		{Code: `
      function _(): string | never {
        return '';
      }
    `},
		{Code: `
      const _ = (): string | never => {
        return '';
      };
    `},
		{Code: `
      type B = string;
      type T = { (): B | never };
    `},
		{Code: "type T = { new (): string | never };"},
		{Code: `
      type B = never;
      type T = { new (): string | B };
    `},
		{Code: `
      type B = unknown;
      type T = B;
    `},
		{Code: "type T = bigint;"},
		{Code: `
      type B = bigint;
      type T = B;
    `},
		{Code: "type T = 1n | 2n;"},
		{Code: `
      type B = 1n;
      type T = B | 2n;
    `},
		{Code: "type T = boolean;"},
		{Code: `
      type B = boolean;
      type T = B;
    `},
		{Code: "type T = false | true;"},
		{Code: `
      type B = false;
      type T = B | true;
    `},
		{Code: `
      type B = true;
      type T = B | false;
    `},
		{Code: "type T = number;"},
		{Code: `
      type B = number;
      type T = B;
    `},
		{Code: "type T = 1 | 2;"},
		{Code: `
      type B = 1;
      type T = B | 2;
    `},
		{Code: "type T = 1 | false;"},
		{Code: `
      type B = 1;
      type T = B | false;
    `},
		{Code: "type T = string;"},
		{Code: `
      type B = string;
      type T = B;
    `},
		{Code: "type T = 'a' | 'b';"},
		{Code: `
      type B = 'b';
      type T = 'a' | B;
    `},
		{Code: `
      type B = 'a';
      type T = B | 'b';
    `},
		{Code: "type T = bigint | null;"},
		{Code: `
      type B = bigint;
      type T = B | null;
    `},
		{Code: "type T = boolean | null;"},
		{Code: `
      type B = boolean;
      type T = B | null;
    `},
		{Code: "type T = number | null;"},
		{Code: `
      type B = number;
      type T = B | null;
    `},
		{Code: "type T = string | null;"},
		{Code: `
      type B = string;
      type T = B | null;
    `},
		{Code: "type T = bigint & null;"},
		{Code: `
      type B = bigint;
      type T = B & null;
    `},
		{Code: "type T = boolean & null;"},
		{Code: `
      type B = boolean;
      type T = B & null;
    `},
		{Code: "type T = number & null;"},
		{Code: `
      type B = number;
      type T = B & null;
    `},
		{Code: "type T = string & null;"},
		{Code: `
      type B = string;
      type T = B & null;
    `},
		{Code: "type T = `${string}` & null;"},
		{Code: `
      type B = ` + "`" + `${string}` + "`" + `;
      type T = B & null;
    `},
		{Code: `
      type T = 'a' | 1 | 'b';
      type U = T & string;
    `},
		{Code: "declare function fn(): never | 'foo';"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "type T = number | any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    19,
				},
			},
		},
		{
			Code: `
        type B = number;
        type T = B | any;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    22,
				},
			},
		},
		{
			Code: "type T = any | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    10,
				},
			},
		},
		{
			Code: `
        type B = any;
        type T = B | number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = number | never;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    19,
				},
			},
		},
		{
			Code: `
        type B = number;
        type T = B | never;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    22,
				},
			},
		},
		{
			Code: `
        type B = never;
        type T = B | number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = never | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = number | unknown;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = unknown | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    10,
				},
			},
		},
		{
			Code: "type ErrorTypes = NotKnown | 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorTypeOverrides",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = number | 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = number | (0 | 1);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = (0 | 0) | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: `
        type B = 0 | 1;
        type T = (2 | B) | number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = (0 | (1 | 2)) | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = (0 | 1) | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = (0 | (0 | 1)) | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = (2 | 'other' | 3) | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = '' | string;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: `
        type B = 'b';
        type T = B | string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = `a${number}c` | string;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: `
        type B = ` + "`" + `a${number}c` + "`" + `;
        type T = B | string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = `${number}` | string;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = 0n | bigint;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = -1n | bigint;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = (-1n | 1n) | bigint;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: `
        type B = boolean;
        type T = B | false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    22,
				},
			},
		},
		{
			Code: "type T = false | boolean;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = true | boolean;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = false & boolean;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    18,
				},
			},
		},
		{
			Code: `
        type B = false;
        type T = B & boolean;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    22,
				},
			},
		},
		{
			Code: `
        type B = true;
        type T = B & boolean;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    22,
				},
			},
		},
		{
			Code: "type T = true & boolean;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    17,
				},
			},
		},
		{
			Code: "type T = number & any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = any & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    10,
				},
			},
		},
		{
			Code: "type ErrorTypes = NotKnown & 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorTypeOverrides",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = number & never;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    19,
				},
			},
		},
		{
			Code: `
        type B = never;
        type T = B & number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    18,
				},
			},
		},
		{
			Code: "type T = never & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = number & unknown;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    19,
				},
			},
		},
		{
			Code: "type T = unknown & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = number & 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    10,
				},
			},
		},
		{
			Code: "type T = '' & string;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    15,
				},
			},
		},
		{
			Code: `
        type B = 0n;
        type T = B & bigint;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    22,
				},
			},
		},
		{
			Code: "type T = 0n & bigint;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    15,
				},
			},
		},
		{
			Code: "type T = -1n & bigint;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    16,
				},
			},
		},
		{
			Code: `
        type T = 'a' | 'b';
        type U = T & string;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    18,
				},
			},
		},
		{
			Code: `
        type S = 1 | 2;
        type T = 'a' | 'b';
        type U = S & T & string & number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    18,
				},
				{
					MessageId: "primitiveOverridden",
					Column:    22,
				},
			},
		},
	})
}
