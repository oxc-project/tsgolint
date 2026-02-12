package consistent_type_exports

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestConsistentTypeExports(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ConsistentTypeExportsRule,
		[]rule_tester.ValidTestCase{
			// unknown module should be ignored
			{Code: "export { Foo } from 'foo';"},
			{Code: "export type { Type1 } from './consistent-type-exports';"},
			{Code: "export { value1 } from './consistent-type-exports';"},
			{Code: "export * from './consistent-type-exports';"},
			{Code: "export type * from './type-only-exports';"},
			{Code: "export * from './value-reexport';"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "export { Type1 } from './consistent-type-exports';",
				Output: []string{"export type { Type1 } from './consistent-type-exports';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: "export { Type1, value1 } from './consistent-type-exports';",
				Output: []string{
					"export type { Type1 } from './consistent-type-exports';\nexport { value1 } from './consistent-type-exports';",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
				},
			},
			{
				Code: "export { Type1, value1, Type2, value2 } from './consistent-type-exports';",
				Output: []string{
					"export type { Type1, Type2 } from './consistent-type-exports';\nexport { value1, value2 } from './consistent-type-exports';",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleExportsAreTypes"},
				},
			},
			{
				Code: `
import { Type2 } from './consistent-type-exports';
export { Type2 };
`,
				Output: []string{`
import { Type2 } from './consistent-type-exports';
export type { Type2 };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
type T = 1;
const x = 1;
export { T, x };
`,
				Options: rule_tester.OptionsFromJSON[ConsistentTypeExportsOptions](`{"fixMixedExportsWithInlineTypeSpecifier": true}`),
				Output: []string{`
type T = 1;
const x = 1;
export { type T, x };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
				},
			},
			{
				Code: `
type T = 1;
export { type T, T };
`,
				Output: []string{`
type T = 1;
export type { T, T };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code:   "export * from './type-only-exports';",
				Output: []string{"export type * from './type-only-exports';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code:   "export * as foo from './type-only-reexport';",
				Output: []string{"export type * as foo from './type-only-reexport';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
		},
	)
}
