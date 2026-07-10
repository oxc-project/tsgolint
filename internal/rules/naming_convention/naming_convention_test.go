package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Port of the upstream static test suite.
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/naming-convention/naming-convention.test.ts
//
// Cases appear in upstream order (invalid first, then valid, as in the source
// file). Where upstream asserts on error `data`, the exact rendered message is
// asserted via the Message field instead.
func TestNamingConvention(t *testing.T) {
	t.Parallel()

	invalidCases := []rule_tester.InvalidTestCase{
		{
			// make sure we handle no options and apply defaults
			Code: "const x_x = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
		},
		{
			// make sure we handle empty options and apply defaults
			Code: "const x_x = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{},
		},
		{
			Code: `
        const child_process = require('child_process');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Filter:   MatchRegex{Match: true, Regex: "child_process"},
					Format:   &[]string{"camelCase"},
					Selector: "default",
				},
			},
		},
		{
			Code: `
        declare const any_camelCase01: any;
        declare const any_camelCase02: any | null;
        declare const any_camelCase03: any | null | undefined;
        declare const string_camelCase01: string;
        declare const string_camelCase02: string | null;
        declare const string_camelCase03: string | null | undefined;
        declare const string_camelCase04: 'a' | null | undefined;
        declare const string_camelCase05: string | 'a' | null | undefined;
        declare const number_camelCase06: number;
        declare const number_camelCase07: number | null;
        declare const number_camelCase08: number | null | undefined;
        declare const number_camelCase09: 1 | null | undefined;
        declare const number_camelCase10: number | 2 | null | undefined;
        declare const boolean_camelCase11: boolean;
        declare const boolean_camelCase12: boolean | null;
        declare const boolean_camelCase13: boolean | null | undefined;
        declare const boolean_camelCase14: true | null | undefined;
        declare const boolean_camelCase15: false | null | undefined;
        declare const boolean_camelCase16: true | false | null | undefined;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
			},
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"const"},
					Prefix:    []string{"any_"},
					Selector:  "variable",
				},
				{
					Format:   &[]string{"snake_case"},
					Prefix:   []string{"string_"},
					Selector: "variable",
					Types:    []string{"string"},
				},
				{
					Format:   &[]string{"snake_case"},
					Prefix:   []string{"number_"},
					Selector: "variable",
					Types:    []string{"number"},
				},
				{
					Format:   &[]string{"snake_case"},
					Prefix:   []string{"boolean_"},
					Selector: "variable",
					Types:    []string{"boolean"},
				},
			},
		},
		{
			Code: `
        declare const function_camelCase1: () => void;
        declare const function_camelCase2: (() => void) | null;
        declare const function_camelCase3: (() => void) | null | undefined;
        declare const function_camelCase4:
          | (() => void)
          | (() => string)
          | null
          | undefined;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Prefix:   []string{"function_"},
					Selector: "variable",
					Types:    []string{"function"},
				},
			},
		},
		{
			Code: `
        declare const array_camelCase1: Array<number>;
        declare const array_camelCase2: ReadonlyArray<number> | null;
        declare const array_camelCase3: number[] | null | undefined;
        declare const array_camelCase4: readonly number[] | null | undefined;
        declare const array_camelCase5:
          | number[]
          | (number | string)[]
          | null
          | undefined;
        declare const array_camelCase6: [] | null | undefined;
        declare const array_camelCase7: [number] | null | undefined;
        declare const array_camelCase8:
          | readonly number[]
          | Array<string>
          | [boolean]
          | null
          | undefined;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Prefix:   []string{"array_"},
					Selector: "variable",
					Types:    []string{"array"},
				},
			},
		},
		{
			Code: `
        let unused_foo = 'a';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Variable name `unused_foo` must not match the RegExp: /^unused_\\w/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:            &MatchRegex{Match: false, Regex: "^unused_\\w"},
					Format:            &[]string{"snake_case"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          "default",
				},
			},
		},
		{
			Code: `
        const _unused_foo = 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Variable name `_unused_foo` must not match the RegExp: /^unused_\\w/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:            &MatchRegex{Match: false, Regex: "^unused_\\w"},
					Format:            &[]string{"snake_case"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          "default",
				},
			},
		},
		{
			Code: `
        interface IFoo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:   &MatchRegex{Match: false, Regex: "^I[A-Z]"},
					Format:   &[]string{"PascalCase"},
					Selector: "typeLike",
				},
			},
		},
		{
			Code: `
        class IBar {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Class name `IBar` must not match the RegExp: /^I[A-Z]/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:   &MatchRegex{Match: false, Regex: "^I[A-Z]"},
					Format:   &[]string{"PascalCase"},
					Selector: "typeLike",
				},
			},
		},
		{
			Code: `
        function fooBar() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Function name `fooBar` must match the RegExp: /function/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:            &MatchRegex{Match: true, Regex: "function"},
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          "function",
				},
			},
		},
		{
			Code: `
        let unused_foo = 'a';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Variable name `unused_foo` must match one of the following formats: camelCase",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          []string{"variable", "function"},
				},
			},
		},
		{
			Code: `
        const _unused_foo = 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormatTrimmed",
					Message:   "Variable name `_unused_foo` trimmed as `unused_foo` must match one of the following formats: camelCase",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          []string{"variable", "function"},
				},
			},
		},
		{
			Code: `
        function foo_bar() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Function name `foo_bar` must match one of the following formats: camelCase",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          []string{"variable", "function"},
				},
			},
		},
		{
			Code: `
        interface IFoo {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Interface name `IFoo` must not match the RegExp: /^I[A-Z]/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Custom:   &MatchRegex{Match: false, Regex: "^I[A-Z]"},
					Format:   &[]string{"PascalCase"},
					Selector: []string{"class", "interface"},
				},
			},
		},
		{
			Code: `
        class IBar {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "satisfyCustom",
					Message:   "Class name `IBar` must not match the RegExp: /^I[A-Z]/u",
					Line:      2,
				},
			},
			Options: []NamingConventionOption{
				{
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          []string{"variable", "function"},
				},
				{
					Custom:   &MatchRegex{Match: false, Regex: "^I[A-Z]"},
					Format:   &[]string{"PascalCase"},
					Selector: []string{"class", "interface"},
				},
			},
		},
		{
			Code: `
        const foo = {
          'Property Name': 'asdf',
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Object Literal Property name `Property Name` must match one of the following formats: strictCamelCase",
					Line:      3,
				},
			},
			Options: []NamingConventionOption{
				{
					Filter:   MatchRegex{Match: false, Regex: "-"},
					Format:   &[]string{"strictCamelCase"},
					Selector: "default",
				},
			},
		},
		{
			Code: `
        const myfoo_bar = 'abcs';
        function fun(myfoo: string) {}
        class foo {
          Myfoo: string;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
				{MessageId: "doesNotMatchFormatTrimmed"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Prefix:   []string{"my", "My"},
					Selector: []string{"variable", "property", "parameter"},
					Types:    []string{"string"},
				},
			},
		},
		{
			Code: `
        class foo {
          private readonly fooBar: boolean;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"private", "readonly"},
					Selector:  []string{"property", "accessor"},
				},
			},
		},
		{
			Code: `
        function my_foo_bar() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormatTrimmed"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Prefix:   []string{"my", "My"},
					Selector: []string{"variable", "function"},
					Types:    []string{"string"},
				},
			},
		},
		{
			Code: `
        class SomeClass {
          static otherConstant = 'hello';
        }

        export const { otherConstant } = SomeClass;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Line:      3,
				},
			},
			Options: []NamingConventionOption{
				{Format: &[]string{"PascalCase"}, Selector: "property"},
				{Format: &[]string{"camelCase"}, Selector: "variable"},
			},
		},
		{
			Code: `
        declare class Foo {
          Bar(Baz: string): void;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Line:      3,
				},
			},
			Options: []NamingConventionOption{
				{Format: &[]string{"camelCase"}, Selector: "parameter"},
			},
		},
		{
			Code: `
        export const PascalCaseVar = 1;
        export enum PascalCaseEnum {}
        export class PascalCaseClass {}
        export function PascalCaseFunction() {}
        export interface PascalCaseInterface {}
        export type PascalCaseType = {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Selector: "default",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "function",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "class",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "interface",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "typeAlias",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "enum",
				},
			},
		},
		{
			Code: `
        const PascalCaseVar = 1;
        enum PascalCaseEnum {}
        class PascalCaseClass {}
        function PascalCaseFunction() {}
        interface PascalCaseInterface {}
        type PascalCaseType = {};
        export {
          PascalCaseVar,
          PascalCaseEnum,
          PascalCaseClass,
          PascalCaseFunction,
          PascalCaseInterface,
          PascalCaseType,
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{Format: &[]string{"snake_case"}, Selector: "default"},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "function",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "class",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "interface",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "typeAlias",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"exported"},
					Selector:  "enum",
				},
			},
		},
		{
			Code: `
        const PascalCaseVar = 1;
        function PascalCaseFunction() {}
        declare function PascalCaseDeclaredFunction();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{Format: &[]string{"snake_case"}, Selector: "default"},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"global"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"global"},
					Selector:  "function",
				},
			},
		},
		{
			Code: `
        const { some_name1 } = {};
        const { some_name2 = 2 } = {};
        const { ignored: IgnoredDueToModifiers1 } = {};
        const { ignored: IgnoredDueToModifiers2 = 3 } = {};
        const IgnoredDueToModifiers3 = 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"destructured"},
					Selector:  "variable",
				},
			},
		},
		{
			Code: `
        export function Foo(
          { aName },
          { anotherName = 1 },
          { ignored: IgnoredDueToModifiers1 },
          { ignored: IgnoredDueToModifiers1 = 2 },
          IgnoredDueToModifiers2,
        ) {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"destructured"},
					Selector:  "parameter",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static abstract readonly some_name;
          IgnoredDueToModifiers = 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"static", "readonly"},
					Selector:  "classProperty",
				},
			},
		},
		{
			Code: `
        class Ignored {
          constructor(
            private readonly some_name,
            IgnoredDueToModifiers,
          ) {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"readonly"},
					Selector:  "parameterProperty",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static some_name() {}
          IgnoredDueToModifiers() {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"static"},
					Selector:  "classMethod",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static get some_name() {}
          get IgnoredDueToModifiers() {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"private", "static"},
					Selector:  "accessor",
				},
			},
		},
		{
			Code: `
        abstract class some_name {}
        class IgnoredDueToModifier {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"abstract"},
					Selector:  "class",
				},
			},
		},
		{
			Code: `
        const UnusedVar = 1;
        function UnusedFunc(
          // this line is intentionally broken out
          UnusedParam: string,
        ) {}
        class UnusedClass {}
        interface UnusedInterface {}
        type UnusedType<
          // this line is intentionally broken out
          UnusedTypeParam,
        > = {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"unused"},
					Selector:  "default",
				},
			},
		},
		{
			Code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Selector: "default",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"requiresQuotes"},
					Selector:  "default",
				},
			},
		},
		{
			Code: `
        type Foo = {
          'foo     Bar': string;
          '': string;
          '0': string;
          'foo': string;
          'foo-bar': string;
          '#foo-bar': string;
        };

        interface Bar {
          'boo-----foo': string;
        }
      `,
			// 6, not 7 because 'foo' is valid
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
				{MessageId: "doesNotMatchFormat"},
			},
		},
		{
			Code: `
        class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
          // ❌ error
          public async asyncBar() {
            return 42;
          }
          // ❌ error
          public AsyncBar2 = async () => {
            return 42;
          };
          // ❌ error
          public AsyncBar3 = async function () {
            return 42;
          };
        }
        abstract class foo {
          public abstract Bar(): number;
          public abstract async async_bar(): number;
          // ❌ error
          public abstract async ASYNC_BAR(): number;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `asyncBar` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `AsyncBar2` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `AsyncBar3` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `ASYNC_BAR` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "method",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"method", "objectLiteralMethod"},
				},
			},
		},
		{
			Code: `
        const obj = {
          Bar() {
            return 42;
          },
          async async_bar() {
            return 42;
          },
          // ❌ error
          async AsyncBar() {
            return 42;
          },
          // ❌ error
          AsyncBar2: async () => {
            return 42;
          },
          // ❌ error
          AsyncBar3: async function () {
            return 42;
          },
        };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Object Literal Method name `AsyncBar` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Object Literal Method name `AsyncBar2` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Object Literal Method name `AsyncBar3` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "method",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"method", "objectLiteralMethod"},
				},
			},
		},
		{
			Code: `
        const syncbar1 = () => {};
        function syncBar2() {}
        const syncBar3 = function syncBar4() {};

        // ❌ error
        const AsyncBar1 = async () => {};
        const async_bar1 = async () => {};
        const async_bar3 = async function async_bar4() {};
        async function async_bar2() {}
        // ❌ error
        const asyncBar5 = async function async_bar6() {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Variable name `AsyncBar1` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Variable name `asyncBar5` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "variableLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"variableLike"},
				},
			},
		},
		{
			Code: `
        const syncbar1 = () => {};
        function syncBar2() {}
        const syncBar3 = function syncBar4() {};

        const async_bar1 = async () => {};
        // ❌ error
        async function asyncBar2() {}
        const async_bar3 = async function async_bar4() {};
        async function async_bar2() {}
        // ❌ error
        const async_bar3 = async function ASYNC_BAR4() {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Function name `asyncBar2` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Function name `ASYNC_BAR4` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "variableLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"variableLike"},
				},
			},
		},
		{
			Code: `
        class foo extends bar {
          public someAttribute = 1;
          public override some_attribute_override = 1;
          // ❌ error
          public override someAttributeOverride = 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Property name `someAttributeOverride` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"override"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: `
        class foo extends bar {
          public override some_method_override() {
            return 42;
          }
          // ❌ error
          public override someMethodOverride() {
            return 42;
          }
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `someMethodOverride` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"override"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: `
        class foo extends bar {
          public get someGetter(): string;
          public override get some_getter_override(): string;
          // ❌ error
          public override get someGetterOverride(): string;
          public set someSetter(val: string);
          public override set some_setter_override(val: string);
          // ❌ error
          public override set someSetterOverride(val: string);
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Classic Accessor name `someGetterOverride` must match one of the following formats: snake_case",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Classic Accessor name `someSetterOverride` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"override"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: `
        class foo {
          private firstPrivateField = 1;
          // ❌ error
          private first_private_field = 1;
          // ❌ error
          #secondPrivateField = 1;
          #second_private_field = 1;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Property name `first_private_field` must match one of the following formats: camelCase",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Property name `secondPrivateField` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"#private"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: `
        class foo {
          private firstPrivateMethod() {}
          // ❌ error
          private first_private_method() {}
          // ❌ error
          #secondPrivateMethod() {}
          #second_private_method() {}
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `first_private_method` must match one of the following formats: camelCase",
				},
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Class Method name `secondPrivateMethod` must match one of the following formats: snake_case",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"#private"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: "import * as fooBar from 'foo_bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Import name `fooBar` must match one of the following formats: PascalCase",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"namespace"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import FooBar from 'foo_bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Import name `FooBar` must match one of the following formats: camelCase",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"namespace"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import { default as foo_bar } from 'foo_bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Import name `foo_bar` must match one of the following formats: camelCase",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"namespace"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import { \"🍎\" as foo } from 'foo_bar';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "doesNotMatchFormat",
					Message:   "Import name `foo` must match one of the following formats: PascalCase",
				},
			},
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
			},
		},
	}

	validCases := []rule_tester.ValidTestCase{
		{
			Code: `
        const child_process = require('child_process');
      `,
			Options: []NamingConventionOption{
				{
					Filter:   MatchRegex{Match: false, Regex: "child_process"},
					Format:   &[]string{"camelCase"},
					Selector: "default",
				},
			},
		},
		{
			Code: `
        declare const ANY_UPPER_CASE: any;
        declare const ANY_UPPER_CASE: any | null;
        declare const ANY_UPPER_CASE: any | null | undefined;

        declare const string_camelCase: string;
        declare const string_camelCase: string | null;
        declare const string_camelCase: string | null | undefined;
        declare const string_camelCase: 'a' | null | undefined;
        declare const string_camelCase: string | 'a' | null | undefined;

        declare const number_camelCase: number;
        declare const number_camelCase: number | null;
        declare const number_camelCase: number | null | undefined;
        declare const number_camelCase: 1 | null | undefined;
        declare const number_camelCase: number | 2 | null | undefined;

        declare const boolean_camelCase: boolean;
        declare const boolean_camelCase: boolean | null;
        declare const boolean_camelCase: boolean | null | undefined;
        declare const boolean_camelCase: true | null | undefined;
        declare const boolean_camelCase: false | null | undefined;
        declare const boolean_camelCase: true | false | null | undefined;
      `,
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"const"},
					Prefix:    []string{"ANY_"},
					Selector:  "variable",
				},
				{
					Format:   &[]string{"camelCase"},
					Prefix:   []string{"string_"},
					Selector: "variable",
					Types:    []string{"string"},
				},
				{
					Format:   &[]string{"camelCase"},
					Prefix:   []string{"number_"},
					Selector: "variable",
					Types:    []string{"number"},
				},
				{
					Format:   &[]string{"camelCase"},
					Prefix:   []string{"boolean_"},
					Selector: "variable",
					Types:    []string{"boolean"},
				},
			},
		},
		{
			Code: `
        let foo = 'a';
        const _foo = 1;
        interface Foo {}
        class Bar {}
        function foo_function_bar() {}
      `,
			Options: []NamingConventionOption{
				{
					Custom:            &MatchRegex{Match: false, Regex: "^unused_\\w"},
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          "default",
				},
				{
					Custom:   &MatchRegex{Match: false, Regex: "^I[A-Z]"},
					Format:   &[]string{"PascalCase"},
					Selector: "typeLike",
				},
				{
					Custom:            &MatchRegex{Match: true, Regex: "_function_"},
					Format:            &[]string{"snake_case"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          "function",
				},
			},
		},
		{
			Code: `
        let foo = 'a';
        const _foo = 1;
        interface foo {}
        class bar {}
        function fooFunctionBar() {}
        function _fooFunctionBar() {}
      `,
			Options: []NamingConventionOption{
				{
					Custom:            &MatchRegex{Match: false, Regex: "^unused_\\w"},
					Format:            &[]string{"camelCase"},
					LeadingUnderscore: strPtr("allow"),
					Selector:          []string{"default", "typeLike", "function"},
				},
			},
		},
		{
			Code: `
        const match = 'test'.match(/test/);
        const [, key, value] = match;
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "default",
				},
			},
		},
		// no format selector
		{
			Code: "const snake_case = 1;",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "default",
				},
				{
					Format:   nil,
					Selector: "variable",
				},
			},
		},
		{
			Code: "const snake_case = 1;",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "default",
				},
				{
					Format:   &[]string{},
					Selector: "variable",
				},
			},
		},
		// https://github.com/typescript-eslint/typescript-eslint/issues/1478
		{
			Code: `
        const child_process = require('child_process');
      `,
			Options: []NamingConventionOption{
				{Format: &[]string{"camelCase", "UPPER_CASE"}, Selector: "variable"},
				{
					Filter:   "child_process",
					Format:   &[]string{"snake_case"},
					Selector: "variable",
				},
			},
		},
		{
			Code: `
        const foo = {
          'Property-Name': 'asdf',
        };
      `,
			Options: []NamingConventionOption{
				{
					Filter:   MatchRegex{Match: false, Regex: "-"},
					Format:   &[]string{"strictCamelCase"},
					Selector: "default",
				},
			},
		},
		{
			Code: `
        const foo = {
          'Property-Name': 'asdf',
        };
      `,
			Options: []NamingConventionOption{
				{
					Filter:   MatchRegex{Match: false, Regex: "^(Property-Name)$"},
					Format:   &[]string{"strictCamelCase"},
					Selector: "default",
				},
			},
		},
		{
			Code: `
        let isFoo = 1;
        class foo {
          shouldBoo: number;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Prefix:   []string{"is", "should", "has", "can", "did", "will"},
					Selector: []string{"variable", "parameter", "property", "accessor"},
					Types:    []string{"number"},
				},
			},
		},
		{
			Code: `
        class foo {
          private readonly FooBoo: boolean;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"private", "readonly"},
					Selector:  []string{"property", "accessor"},
					Types:     []string{"boolean"},
				},
			},
		},
		{
			Code: `
        class foo {
          private fooBoo: number;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"private"},
					Selector:  []string{"property", "accessor"},
				},
			},
		},
		{
			Code: `
        const isfooBar = 1;
        function fun(goodfunFoo: number) {}
        class foo {
          private VanFooBar: number;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"StrictPascalCase"},
					Modifiers: []string{"private"},
					Prefix:    []string{"Van"},
					Selector:  []string{"property", "accessor"},
				},
				{
					Format:   &[]string{"camelCase"},
					Prefix:   []string{"is", "good"},
					Selector: []string{"variable", "parameter"},
					Types:    []string{"number"},
				},
			},
		},
		{
			Code: `
        class SomeClass {
          static OtherConstant = 'hello';
        }

        export const { OtherConstant: otherConstant } = SomeClass;
      `,
			Options: []NamingConventionOption{
				{Format: &[]string{"PascalCase"}, Selector: "property"},
				{Format: &[]string{"camelCase"}, Selector: "variable"},
			},
		},
		// treat properties with function expressions as typeMethod
		{
			Code: `
        interface SOME_INTERFACE {
          SomeMethod: () => void;

          some_property: string;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"UPPER_CASE"},
					Selector: "default",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "typeMethod",
				},
				{
					Format:   &[]string{"snake_case"},
					Selector: "typeProperty",
				},
			},
		},
		{
			Code: `
        type Ignored = {
          ignored_due_to_modifiers: string;
          readonly FOO: string;
        };
      `,
			Options: []NamingConventionOption{
				{
					Format:    &[]string{"UPPER_CASE"},
					Modifiers: []string{"readonly"},
					Selector:  "typeProperty",
				},
			},
		},
		{
			Code: `
        const camelCaseVar = 1;
        enum camelCaseEnum {}
        class camelCaseClass {}
        function camelCaseFunction() {}
        interface camelCaseInterface {}
        type camelCaseType = {};
        export const PascalCaseVar = 1;
        export enum PascalCaseEnum {}
        export class PascalCaseClass {}
        export function PascalCaseFunction() {}
        export interface PascalCaseInterface {}
        export type PascalCaseType = {};
      `,
			Options: []NamingConventionOption{
				{Format: &[]string{"camelCase"}, Selector: "default"},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "function",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "class",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "interface",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "typeAlias",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "enum",
				},
			},
		},
		{
			Code: `
        const camelCaseVar = 1;
        enum camelCaseEnum {}
        class camelCaseClass {}
        function camelCaseFunction() {}
        interface camelCaseInterface {}
        type camelCaseType = {};
        const PascalCaseVar = 1;
        enum PascalCaseEnum {}
        class PascalCaseClass {}
        function PascalCaseFunction() {}
        interface PascalCaseInterface {}
        type PascalCaseType = {};
        export {
          PascalCaseVar,
          PascalCaseEnum,
          PascalCaseClass,
          PascalCaseFunction,
          PascalCaseInterface,
          PascalCaseType,
        };
      `,
			Options: []NamingConventionOption{
				{Format: &[]string{"camelCase"}, Selector: "default"},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "function",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "class",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "interface",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "typeAlias",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"exported"},
					Selector:  "enum",
				},
			},
		},
		{
			Code: `
        {
          const camelCaseVar = 1;
          function camelCaseFunction() {}
          declare function camelCaseDeclaredFunction();
        }
        const PascalCaseVar = 1;
        function PascalCaseFunction() {}
        declare function PascalCaseDeclaredFunction();
      `,
			Options: []NamingConventionOption{
				{Format: &[]string{"camelCase"}, Selector: "default"},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"global"},
					Selector:  "variable",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"global"},
					Selector:  "function",
				},
			},
		},
		{
			Code: `
        const { some_name1 } = {};
        const { ignore: IgnoredDueToModifiers1 } = {};
        const { some_name2 = 2 } = {};
        const IgnoredDueToModifiers2 = 1;
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"destructured"},
					Selector:  "variable",
				},
			},
		},
		{
			Code: `
        const { some_name1 } = {};
        const { ignore: IgnoredDueToModifiers1 } = {};
        const { some_name2 = 2 } = {};
        const IgnoredDueToModifiers2 = 1;
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    nil,
					Modifiers: []string{"destructured"},
					Selector:  "variable",
				},
			},
		},
		{
			Code: `
        export function Foo(
          { aName },
          { anotherName = 1 },
          { ignored: IgnoredDueToModifiers1 },
          { ignored: IgnoredDueToModifiers1 = 2 },
          IgnoredDueToModifiers2,
        ) {}
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"destructured"},
					Selector:  "parameter",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static abstract readonly some_name;
          IgnoredDueToModifiers = 1;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"static", "readonly"},
					Selector:  "classProperty",
				},
			},
		},
		{
			Code: `
        class Ignored {
          constructor(
            private readonly some_name,
            IgnoredDueToModifiers,
          ) {}
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"readonly"},
					Selector:  "parameterProperty",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static some_name() {}
          IgnoredDueToModifiers() {}
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"static"},
					Selector:  "classMethod",
				},
			},
		},
		{
			Code: `
        class Ignored {
          private static get some_name() {}
          get IgnoredDueToModifiers() {}
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"private", "static"},
					Selector:  "accessor",
				},
			},
		},
		{
			Code: `
        abstract class some_name {}
        class IgnoredDueToModifier {}
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: "default",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"abstract"},
					Selector:  "class",
				},
			},
		},
		{
			Code: `
        const UnusedVar = 1;
        function UnusedFunc(
          // this line is intentionally broken out
          UnusedParam: string,
        ) {}
        class UnusedClass {}
        interface UnusedInterface {}
        type UnusedType<
          // this line is intentionally broken out
          UnusedTypeParam,
        > = {};

        export const used_var = 1;
        export function used_func(
          // this line is intentionally broken out
          used_param: string,
        ) {
          return used_param;
        }
        export class used_class {}
        export interface used_interface {}
        export type used_type<
          // this line is intentionally broken out
          used_typeparam,
        > = used_typeparam;
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Selector: "default",
				},
				{
					Format:    &[]string{"PascalCase"},
					Modifiers: []string{"unused"},
					Selector:  "default",
				},
			},
		},
		{
			Code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Selector: "default",
				},
				{
					Format:    nil,
					Modifiers: []string{"requiresQuotes"},
					Selector:  "default",
				},
			},
		},
		{
			Code: `
        const ignored1 = {
          'a a': 1,
          'b b'() {},
          get 'c c'() {
            return 1;
          },
          set 'd d'(value: string) {},
        };
        class ignored2 {
          'a a' = 1;
          'b b'() {}
          get 'c c'() {
            return 1;
          }
          set 'd d'(value: string) {}
        }
        interface ignored3 {
          'a a': 1;
          'b b'(): void;
        }
        type ignored4 = {
          'a a': 1;
          'b b'(): void;
        };
        enum ignored5 {
          'a a',
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"snake_case"},
					Selector: "default",
				},
				{
					Format:    nil,
					Modifiers: []string{"requiresQuotes"},
					Selector: []string{
						"classProperty",
						"objectLiteralProperty",
						"typeProperty",
						"classMethod",
						"objectLiteralMethod",
						"typeMethod",
						"accessor",
						"enumMember",
					},
				},
				// making sure the `requiresQuotes` modifier appropriately overrides this
				{
					Format: &[]string{"PascalCase"},
					Selector: []string{
						"classProperty",
						"objectLiteralProperty",
						"typeProperty",
						"classMethod",
						"objectLiteralMethod",
						"typeMethod",
						"accessor",
						"enumMember",
					},
				},
			},
		},
		{
			Code: `
        const obj = {
          Foo: 42,
          Bar() {
            return 42;
          },
        };
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "property",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "method",
				},
			},
		},
		{
			Code: `
        const obj = {
          Bar() {
            return 42;
          },
          async async_bar() {
            return 42;
          },
        };
        class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
        }
        abstract class foo {
          public Bar() {
            return 42;
          }
          public async async_bar() {
            return 42;
          }
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"method", "objectLiteralMethod"},
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "method",
				},
			},
		},
		{
			Code: `
        const async_bar1 = async () => {};
        async function async_bar2() {}
        const async_bar3 = async function async_bar4() {};
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:   &[]string{"PascalCase"},
					Selector: "method",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"async"},
					Selector:  []string{"variable"},
				},
			},
		},
		{
			Code: `
        class foo extends bar {
          public someAttribute = 1;
          public override some_attribute_override = 1;
          public someMethod() {
            return 42;
          }
          public override some_method_override2() {
            return 42;
          }
        }
        abstract class foo extends bar {
          public abstract someAttribute: string;
          public abstract override some_attribute_override: string;
          public abstract someMethod(): string;
          public abstract override some_method_override2(): string;
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"override"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: `
        class foo {
          private someAttribute = 1;
          #some_attribute = 1;

          private someMethod() {}
          #some_method() {}
        }
      `,
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"camelCase"},
					Selector: "memberLike",
				},
				{
					Format:    &[]string{"snake_case"},
					Modifiers: []string{"#private"},
					Selector:  []string{"memberLike"},
				},
			},
		},
		{
			Code: "import * as FooBar from 'foo_bar';",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"default"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import fooBar from 'foo_bar';",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"default"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import { default as fooBar } from 'foo_bar';",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"default"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import { foo_bar } from 'foo_bar';",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
				{
					Format:    &[]string{"camelCase"},
					Modifiers: []string{"default"},
					Selector:  []string{"import"},
				},
			},
		},
		{
			Code: "import { \"🍎\" as Foo } from 'foo_bar';",
			Options: []NamingConventionOption{
				{
					Format:   &[]string{"PascalCase"},
					Selector: []string{"import"},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, validCases, invalidCases)
}
