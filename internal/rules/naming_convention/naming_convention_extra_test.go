package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Additional cases covering behaviour that is easy to get wrong when working on
// the TypeScript AST instead of ESTree.
func TestNamingConventionRuleExtra(t *testing.T) {
	t.Parallel()

	unusedOptions := rule_tester.OptionsFromJSON[NamingConventionOptions](`[
		{"selector": "default", "format": ["PascalCase"]},
		{"selector": "default", "modifiers": ["unused"], "format": ["snake_case"]}
	]`)
	exportedOptions := rule_tester.OptionsFromJSON[NamingConventionOptions](`[
		{"selector": "default", "format": ["snake_case"]},
		{"selector": ["variable", "class", "function"], "modifiers": ["exported"], "format": ["PascalCase"]}
	]`)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			// self references don't count as usages
			{
				Code: `
function unused_recursive() {
  unused_recursive();
}
class unused_class {
  static Create() {
    return new unused_class();
  }
}
interface unused_interface {
  Next: unused_interface;
}
type unused_type = unused_type[];
const unused_arrow = () => unused_arrow();
export const Foo = class unused_class_expression {};
let unused_write_only;
unused_write_only = 1;
`,
				Options: unusedOptions,
			},
			// usages through shorthand properties, types and exports
			{
				Code: `
const UsedVar = 1;
const UsedShorthand = { UsedVar };
interface UsedInterface {}
export const UsedExport: UsedInterface = UsedShorthand;
export default function UsedFunction() {}
declare const AmbientVar: string;
`,
				Options: unusedOptions,
			},
			// `export default` and renamed exports mark declarations as exported
			{
				Code: `
const DefaultExport = 1;
export default DefaultExport;
class RenamedExport {}
export { RenamedExport as renamed_export };
function not_exported() {}
`,
				Options: exportedOptions,
			},
			// `export =` is not considered an export
			{
				Code: `
const export_equals = 1;
export = export_equals;
`,
				Options: exportedOptions,
			},
			// object literals used as destructuring assignment targets are patterns
			{
				Code: `
declare const obj: { snake_prop: number; nested: { inner_prop: number } };
let a: number;
let b: number;
({ snake_prop: a } = obj);
({ nested: { inner_prop: b } } = obj);
[{ snake_prop: a }] = [obj];
for ({ snake_prop: a } of [obj]) {}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"selector": "objectLiteralProperty", "format": ["PascalCase"]}]`),
			},
			// parameters of function types and signatures are not checked
			{
				Code: `
type Fn = (snake_param: string) => void;
type Ctor = new (snake_param: string) => object;
interface Iface {
  method(snake_param: string): void;
  (snake_call: string): void;
}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"selector": "parameter", "format": ["PascalCase"]}]`),
			},
			// catch clause variables are not variables
			{
				Code: `
try {
} catch (snake_error) {}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"selector": "variable", "format": ["PascalCase"]}]`),
			},
			// `infer` type parameters and mapped type keys are not type parameters
			{
				Code: `
type Unwrap<T> = T extends Promise<infer snake_u> ? snake_u : T;
type Mapped<T> = { [snake_k in keyof T]: T[snake_k] };
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"selector": "typeParameter", "format": ["PascalCase"]}]`),
			},
			// accessor signatures are type methods
			{
				Code: `
interface iface {
  get PascalGetter(): number;
  set PascalSetter(value: number);
}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[
					{"selector": "default", "format": ["camelCase"]},
					{"selector": "typeMethod", "format": ["PascalCase"]}
				]`),
			},
			// `var` declarations in top-level `for` heads are global, `let`/`const` are not
			{
				Code: `
for (var GlobalVar = 0; GlobalVar < 1; GlobalVar++) {}
for (const local_const of []) {}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[
					{"selector": "variable", "format": ["snake_case"]},
					{"selector": "variable", "modifiers": ["global"], "format": ["PascalCase"]}
				]`),
			},
			// a single selector object is accepted instead of an array
			{
				Code:    `const snake_case = 1;`,
				Options: map[string]any{"selector": "variable", "format": []any{"snake_case"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// self references don't count as usages
			{
				Code: `
function UnusedRecursive() {
  UnusedRecursive();
}
`,
				Options: unusedOptions,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 10, EndColumn: 25},
				},
			},
			// `export default` marks declarations as exported
			{
				Code: `
const snake_default = 1;
export default snake_default;
`,
				Options: exportedOptions,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 7, EndColumn: 20},
				},
			},
			// `await using` declarations are not `const`
			{
				Code: `
async function foo() {
  await using snake_resource = null as any;
  using other_resource = null as any;
}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[
					{"selector": "variable", "format": ["PascalCase"]},
					{"selector": "variable", "modifiers": ["const"], "format": ["snake_case"]}
				]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 15, EndColumn: 29},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 9, EndColumn: 23},
				},
			},
			// the reported range of parameters includes `?` and the type annotation
			{
				Code: `
function foo(Snake_Param?: string, Other_Param?) {}
class Foo {
  constructor(private readonly Snake_Property: string) {}
}
`,
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"selector": ["parameter", "parameterProperty"], "format": ["camelCase"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 14, EndColumn: 34},
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 36, EndColumn: 48},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 32, EndColumn: 54},
				},
			},
			// computed string literal enum members are validated
			{
				Code: `
enum Foo {
  ['snake_member'] = 1,
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 4, EndColumn: 18},
				},
			},
			// private identifiers are reported including the `#`
			{
				Code: `
class Foo {
  #snake_field = 1;
  #snake_method() {}
  accessor #snake_accessor = 1;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 3, EndColumn: 15},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 3, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 12, EndColumn: 27},
				},
			},
		},
	)
}
