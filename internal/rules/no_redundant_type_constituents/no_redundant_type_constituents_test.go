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
					Column:    10,
					EndColumn: 16,
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
					Column:    18,
					EndColumn: 19,
				},
			},
		},
		{
			Code: "type T = any | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    16,
					EndColumn: 22,
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
					Column:    22,
					EndColumn: 28,
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
					Column:    10,
					EndColumn: 16,
				},
			},
		},
		{
			Code: "type T = unknown | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    20,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "type ErrorTypes = NotKnown | 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorTypeOverrides",
					Column:    30,
					EndColumn: 31,
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
					Column:    10,
					EndColumn: 16,
				},
			},
		},
		{
			Code: "type T = any & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    16,
					EndColumn: 22,
				},
			},
		},
		{
			Code: "type ErrorTypes = NotKnown & 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "errorTypeOverrides",
					Column:    30,
					EndColumn: 31,
				},
			},
		},
		{
			Code: "type T = number & never;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    10,
					EndColumn: 16,
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
					Column:    22,
					EndColumn: 28,
				},
			},
		},
		{
			Code: "type T = never & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "overrides",
					Column:    18,
					EndColumn: 24,
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
					Column:    22,
					EndColumn: 28,
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
					Column:    35,
					EndColumn: 41,
				},
				{
					MessageId: "primitiveOverridden",
					Column:    26,
					EndColumn: 32,
				},
			},
		},
		{
			// Only the numeric member of this constituent is redundant. The string
			// member must not be included in the redundant-type label.
			Code: "type T = (2 | 'other') | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "literalOverridden",
					Column:    10,
					EndColumn: 23,
				},
			},
		},
		{
			// Each literal that overrides the primitive needs its own labeled range.
			Code: "type T = 0 & 1 & number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "primitiveOverridden",
					Column:    18,
					EndColumn: 24,
				},
			},
		},
		{
			Code: "type T = number | string | any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "overrides"},
			},
		},
		{
			Code: "type T = number | string | never;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "overridden"},
			},
		},
		{
			Code: "type T = number & string & never;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "overrides"},
			},
		},
		{
			Code: "type T = number & string & unknown;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "overridden"},
			},
		},
		{
			// The overriding label must point at the nested top type, not its group.
			Code: "type T = (number | any) | string;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "overrides"},
				{MessageId: "overrides"},
			},
		},
		{
			// Preserve the signed literal node when only part of a group is redundant.
			Code: "type T = (-1 | 'other') | number;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalOverridden"},
			},
		},
		{
			// Checker-derived union parts should fall back to their alias subnode,
			// not the entire parenthesized constituent.
			Code: `
        type B = 0 | 'other';
        type T = (2 | B) | number;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalOverridden"},
			},
		},
		{
			// Every matching primitive constituent is redundant in the intersection.
			Code: `
        type N = number;
        type T = (0 | 1) & number & N;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "primitiveOverridden"},
			},
		},
		{
			// The primitive label must point at the matching nested member, not its group.
			Code: "type T = 0 | (number | string);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "literalOverridden"},
			},
		},
	})
}
