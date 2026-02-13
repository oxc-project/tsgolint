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
			{Code: "export { value1 as \"🍎\" } from './consistent-type-exports';"},
			{Code: "export type { value1 } from './consistent-type-exports';"},
			{Code: `
const variable = 1;
class Class {}
enum Enum {}
function Func() {}
namespace ValueNS {
	export const x = 1;
}

export { variable, Class, Enum, Func, ValueNS };
`},
			{Code: `
type Alias = 1;
interface IFace {}
namespace TypeNS {
	export type x = 1;
}

export type { Alias, IFace, TypeNS };
`},
			{Code: `
const foo = 1;
export type { foo };
`},
			{Code: `
namespace NonTypeNS {
	export const x = 1;
}

export { NonTypeNS };
`},
			{Code: "export * from './unknown-module';"},
			{Code: "export * from './consistent-type-exports';"},
			{Code: "export type * from './type-only-exports';"},
			{Code: "export type * from './type-only-reexport';"},
			{Code: "export * from './value-reexport';"},
			{Code: "export * as foo from './consistent-type-exports';"},
			{Code: "export type * as foo from './type-only-exports';"},
			{Code: "export type * as foo from './type-only-reexport';"},
			{Code: "export * as foo from './value-reexport';"},
			{Code: `
import * as Foo from './consistent-type-exports';
type Foo = 1;
export { Foo };
`},
			{Code: `
import { Type1 } from './consistent-type-exports';
const Type1 = 1;
export { Type1 };
`},
			{Code: `
export { A } from './reexport-2-named';
`},
			{Code: `
import { A } from './reexport-2-named';
export { A };
`},
			{Code: `
export { A } from './reexport-2-namespace';
`},
			{Code: `
import { A } from './reexport-2-namespace';
export { A };
`},
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
				Code:   "export { Type1 as \"🍎\" } from './consistent-type-exports';",
				Output: []string{"export type { Type1 as \"🍎\" } from './consistent-type-exports';"},
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
				Code: `
export { Type1, value1, value2 } from './consistent-type-exports';
`,
				Output: []string{`
export type { Type1 } from './consistent-type-exports';
export { value1, value2 } from './consistent-type-exports';
`},
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
				Code:   "export { Type2 as Foo } from './consistent-type-exports';",
				Output: []string{"export type { Type2 as Foo } from './consistent-type-exports';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
export { Type2 as Foo, value1 } from './consistent-type-exports';
`,
				Output: []string{`
export type { Type2 as Foo } from './consistent-type-exports';
export { value1 } from './consistent-type-exports';
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
				},
			},
			{
				Code: `
export {
	Type2 as Foo,
	value1 as BScope,
	value2 as CScope,
} from './consistent-type-exports';
`,
				Output: []string{`
export type { Type2 as Foo } from './consistent-type-exports';
export { value1 as BScope, value2 as CScope } from './consistent-type-exports';
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
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
import { value2, Type2 } from './consistent-type-exports';
export { value2, Type2 };
`,
				Output: []string{`
import { value2, Type2 } from './consistent-type-exports';
export type { Type2 };
export { value2 };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
				},
			},
			{
				Code: `
type Alias = 1;
interface IFace {}
namespace TypeNS {
	export type x = 1;
	export const f = 1;
}

export { Alias, IFace, TypeNS };
`,
				Output: []string{`
type Alias = 1;
interface IFace {}
namespace TypeNS {
	export type x = 1;
	export const f = 1;
}

export type { Alias, IFace };
export { TypeNS };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleExportsAreTypes"},
				},
			},
			{
				Code: `
namespace TypeNS {
	export interface Foo {}
}

export { TypeNS };
`,
				Output: []string{`
namespace TypeNS {
	export interface Foo {}
}

export type { TypeNS };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
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
				Code: `
type T = 1;
export { type/* */T, type     /* */T, T };
`,
				Output: []string{`
type T = 1;
export type { /* */T, /* */T, T };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
type T = 1;
const x = 1;
export { type T, T, x };
`,
				Output: []string{`
type T = 1;
const x = 1;
export type { T, T };
export { x };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "singleExportIsType"},
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
				Options: rule_tester.OptionsFromJSON[ConsistentTypeExportsOptions](`{"fixMixedExportsWithInlineTypeSpecifier": true}`),
				Output: []string{`
type T = 1;
export type { T, T };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
export {
	Type1,
	Type2 as Foo,
	type value1 as BScope,
	value2 as CScope,
} from './consistent-type-exports';
`,
				Options: rule_tester.OptionsFromJSON[ConsistentTypeExportsOptions](`{"fixMixedExportsWithInlineTypeSpecifier": false}`),
				Output: []string{`
export type { Type1, Type2 as Foo, value1 as BScope } from './consistent-type-exports';
export { value2 as CScope } from './consistent-type-exports';
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleExportsAreTypes"},
				},
			},
			{
				Code: `
export {
	Type1,
	Type2 as Foo,
	type value1 as BScope,
	value2 as CScope,
} from './consistent-type-exports';
`,
				Options: rule_tester.OptionsFromJSON[ConsistentTypeExportsOptions](`{"fixMixedExportsWithInlineTypeSpecifier": true}`),
				Output: []string{`
export {
	type Type1,
	type Type2 as Foo,
	type value1 as BScope,
	value2 as CScope,
} from './consistent-type-exports';
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleExportsAreTypes"},
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
				Code:   "export * from './type-only-reexport';",
				Output: []string{"export type * from './type-only-reexport';"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
/* comment 1 */ export
	/* comment 2 */ *
		// comment 3
		from './type-only-exports';
`,
				Output: []string{`
/* comment 1 */ export
	/* comment 2 */ type *
		// comment 3
		from './type-only-exports';
`},
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
			{
				Code: `
import type * as Foo from './consistent-type-exports';
type Foo = 1;
export { Foo };
`,
				Output: []string{`
import type * as Foo from './consistent-type-exports';
type Foo = 1;
export type { Foo };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
			{
				Code: `
import { type NAME as Foo } from './consistent-type-exports';
export { Foo };
`,
				Output: []string{`
import { type NAME as Foo } from './consistent-type-exports';
export type { Foo };
`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "typeOverValue"},
				},
			},
		},
	)
}
