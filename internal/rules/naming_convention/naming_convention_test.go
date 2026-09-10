package naming_convention

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNamingConventionRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NamingConventionRule,
		[]rule_tester.ValidTestCase{
			{Code: "\n        const child_process = require('child_process');\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"filter":{"match":false,"regex":"child_process"},"format":["camelCase"],"selector":"default"}]`)},
			{Code: "\n        declare const ANY_UPPER_CASE: any;\n        declare const ANY_UPPER_CASE: any | null;\n        declare const ANY_UPPER_CASE: any | null | undefined;\n\n        declare const string_camelCase: string;\n        declare const string_camelCase: string | null;\n        declare const string_camelCase: string | null | undefined;\n        declare const string_camelCase: 'a' | null | undefined;\n        declare const string_camelCase: string | 'a' | null | undefined;\n\n        declare const number_camelCase: number;\n        declare const number_camelCase: number | null;\n        declare const number_camelCase: number | null | undefined;\n        declare const number_camelCase: 1 | null | undefined;\n        declare const number_camelCase: number | 2 | null | undefined;\n\n        declare const boolean_camelCase: boolean;\n        declare const boolean_camelCase: boolean | null;\n        declare const boolean_camelCase: boolean | null | undefined;\n        declare const boolean_camelCase: true | null | undefined;\n        declare const boolean_camelCase: false | null | undefined;\n        declare const boolean_camelCase: true | false | null | undefined;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["UPPER_CASE"],"modifiers":["const"],"prefix":["ANY_"],"selector":"variable"},{"format":["camelCase"],"prefix":["string_"],"selector":"variable","types":["string"]},{"format":["camelCase"],"prefix":["number_"],"selector":"variable","types":["number"]},{"format":["camelCase"],"prefix":["boolean_"],"selector":"variable","types":["boolean"]}]`)},
			{Code: "\n        let foo = 'a';\n        const _foo = 1;\n        interface Foo {}\n        class Bar {}\n        function foo_function_bar() {}\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^unused_\\w"},"format":["camelCase"],"leadingUnderscore":"allow","selector":"default"},{"custom":{"match":false,"regex":"^I[A-Z]"},"format":["PascalCase"],"selector":"typeLike"},{"custom":{"match":true,"regex":"_function_"},"format":["snake_case"],"leadingUnderscore":"allow","selector":"function"}]`)},
			{Code: "\n        let foo = 'a';\n        const _foo = 1;\n        interface foo {}\n        class bar {}\n        function fooFunctionBar() {}\n        function _fooFunctionBar() {}\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^unused_\\w"},"format":["camelCase"],"leadingUnderscore":"allow","selector":["default","typeLike","function"]}]`)},
			{Code: "\n        const match = 'test'.match(/test/);\n        const [, key, value] = match;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"}]`)},
			{Code: "const snake_case = 1;", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"},{"format":null,"selector":"variable"}]`)},
			{Code: "const snake_case = 1;", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"},{"format":[],"selector":"variable"}]`)},
			{Code: "\n        const child_process = require('child_process');\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase","UPPER_CASE"],"selector":"variable"},{"filter":"child_process","format":["snake_case"],"selector":"variable"}]`)},
			{Code: "\n        const foo = {\n          'Property-Name': 'asdf',\n        };\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"filter":{"match":false,"regex":"-"},"format":["strictCamelCase"],"selector":"default"}]`)},
			{Code: "\n        const foo = {\n          'Property-Name': 'asdf',\n        };\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"filter":{"match":false,"regex":"^(Property-Name)$"},"format":["strictCamelCase"],"selector":"default"}]`)},
			{Code: "\n        let isFoo = 1;\n        class foo {\n          shouldBoo: number;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"prefix":["is","should","has","can","did","will"],"selector":["variable","parameter","property","accessor"],"types":["number"]}]`)},
			{Code: "\n        class foo {\n          private readonly FooBoo: boolean;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"modifiers":["private","readonly"],"selector":["property","accessor"],"types":["boolean"]}]`)},
			{Code: "\n        class foo {\n          private fooBoo: number;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"modifiers":["private"],"selector":["property","accessor"]}]`)},
			{Code: "\n        const isfooBar = 1;\n        function fun(goodfunFoo: number) {}\n        class foo {\n          private VanFooBar: number;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["StrictPascalCase"],"modifiers":["private"],"prefix":["Van"],"selector":["property","accessor"]},{"format":["camelCase"],"prefix":["is","good"],"selector":["variable","parameter"],"types":["number"]}]`)},
			{Code: "\n        class SomeClass {\n          static OtherConstant = 'hello';\n        }\n\n        export const { OtherConstant: otherConstant } = SomeClass;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"property"},{"format":["camelCase"],"selector":"variable"}]`)},
			{Code: "\n        interface SOME_INTERFACE {\n          SomeMethod: () => void;\n\n          some_property: string;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["UPPER_CASE"],"selector":"default"},{"format":["PascalCase"],"selector":"typeMethod"},{"format":["snake_case"],"selector":"typeProperty"}]`)},
			{Code: "\n        type Ignored = {\n          ignored_due_to_modifiers: string;\n          readonly FOO: string;\n        };\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["UPPER_CASE"],"modifiers":["readonly"],"selector":"typeProperty"}]`)},
			{Code: "\n        const camelCaseVar = 1;\n        enum camelCaseEnum {}\n        class camelCaseClass {}\n        function camelCaseFunction() {}\n        interface camelCaseInterface {}\n        type camelCaseType = {};\n        export const PascalCaseVar = 1;\n        export enum PascalCaseEnum {}\n        export class PascalCaseClass {}\n        export function PascalCaseFunction() {}\n        export interface PascalCaseInterface {}\n        export type PascalCaseType = {};\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"variable"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"function"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"class"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"interface"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"typeAlias"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"enum"}]`)},
			{Code: "\n        const camelCaseVar = 1;\n        enum camelCaseEnum {}\n        class camelCaseClass {}\n        function camelCaseFunction() {}\n        interface camelCaseInterface {}\n        type camelCaseType = {};\n        const PascalCaseVar = 1;\n        enum PascalCaseEnum {}\n        class PascalCaseClass {}\n        function PascalCaseFunction() {}\n        interface PascalCaseInterface {}\n        type PascalCaseType = {};\n        export {\n          PascalCaseVar,\n          PascalCaseEnum,\n          PascalCaseClass,\n          PascalCaseFunction,\n          PascalCaseInterface,\n          PascalCaseType,\n        };\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"variable"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"function"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"class"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"interface"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"typeAlias"},{"format":["PascalCase"],"modifiers":["exported"],"selector":"enum"}]`)},
			{Code: "\n        {\n          const camelCaseVar = 1;\n          function camelCaseFunction() {}\n          declare function camelCaseDeclaredFunction();\n        }\n        const PascalCaseVar = 1;\n        function PascalCaseFunction() {}\n        declare function PascalCaseDeclaredFunction();\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"default"},{"format":["PascalCase"],"modifiers":["global"],"selector":"variable"},{"format":["PascalCase"],"modifiers":["global"],"selector":"function"}]`)},
			{Code: "\n        const { some_name1 } = {};\n        const { ignore: IgnoredDueToModifiers1 } = {};\n        const { some_name2 = 2 } = {};\n        const IgnoredDueToModifiers2 = 1;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["destructured"],"selector":"variable"}]`)},
			{Code: "\n        const { some_name1 } = {};\n        const { ignore: IgnoredDueToModifiers1 } = {};\n        const { some_name2 = 2 } = {};\n        const IgnoredDueToModifiers2 = 1;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":null,"modifiers":["destructured"],"selector":"variable"}]`)},
			{Code: "\n        export function Foo(\n          { aName },\n          { anotherName = 1 },\n          { ignored: IgnoredDueToModifiers1 },\n          { ignored: IgnoredDueToModifiers1 = 2 },\n          IgnoredDueToModifiers2,\n        ) {}\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["camelCase"],"modifiers":["destructured"],"selector":"parameter"}]`)},
			{Code: "\n        class Ignored {\n          private static abstract readonly some_name;\n          IgnoredDueToModifiers = 1;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["static","readonly"],"selector":"classProperty"}]`)},
			{Code: "\n        class Ignored {\n          constructor(\n            private readonly some_name,\n            IgnoredDueToModifiers,\n          ) {}\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["readonly"],"selector":"parameterProperty"}]`)},
			{Code: "\n        class Ignored {\n          private static some_name() {}\n          IgnoredDueToModifiers() {}\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["static"],"selector":"classMethod"}]`)},
			{Code: "\n        class Ignored {\n          private static get some_name() {}\n          get IgnoredDueToModifiers() {}\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["private","static"],"selector":"accessor"}]`)},
			{Code: "\n        abstract class some_name {}\n        class IgnoredDueToModifier {}\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["abstract"],"selector":"class"}]`)},
			{Code: "\n        const UnusedVar = 1;\n        function UnusedFunc(\n          // this line is intentionally broken out\n          UnusedParam: string,\n        ) {}\n        class UnusedClass {}\n        interface UnusedInterface {}\n        type UnusedType<\n          // this line is intentionally broken out\n          UnusedTypeParam,\n        > = {};\n\n        export const used_var = 1;\n        export function used_func(\n          // this line is intentionally broken out\n          used_param: string,\n        ) {\n          return used_param;\n        }\n        export class used_class {}\n        export interface used_interface {}\n        export type used_type<\n          // this line is intentionally broken out\n          used_typeparam,\n        > = used_typeparam;\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":["PascalCase"],"modifiers":["unused"],"selector":"default"}]`)},
			{Code: "\n        const ignored1 = {\n          'a a': 1,\n          'b b'() {},\n          get 'c c'() {\n            return 1;\n          },\n          set 'd d'(value: string) {},\n        };\n        class ignored2 {\n          'a a' = 1;\n          'b b'() {}\n          get 'c c'() {\n            return 1;\n          }\n          set 'd d'(value: string) {}\n        }\n        interface ignored3 {\n          'a a': 1;\n          'b b'(): void;\n        }\n        type ignored4 = {\n          'a a': 1;\n          'b b'(): void;\n        };\n        enum ignored5 {\n          'a a',\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":null,"modifiers":["requiresQuotes"],"selector":"default"}]`)},
			{Code: "\n        const ignored1 = {\n          'a a': 1,\n          'b b'() {},\n          get 'c c'() {\n            return 1;\n          },\n          set 'd d'(value: string) {},\n        };\n        class ignored2 {\n          'a a' = 1;\n          'b b'() {}\n          get 'c c'() {\n            return 1;\n          }\n          set 'd d'(value: string) {}\n        }\n        interface ignored3 {\n          'a a': 1;\n          'b b'(): void;\n        }\n        type ignored4 = {\n          'a a': 1;\n          'b b'(): void;\n        };\n        enum ignored5 {\n          'a a',\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":null,"modifiers":["requiresQuotes"],"selector":["classProperty","objectLiteralProperty","typeProperty","classMethod","objectLiteralMethod","typeMethod","accessor","enumMember"]},{"format":["PascalCase"],"selector":["classProperty","objectLiteralProperty","typeProperty","classMethod","objectLiteralMethod","typeMethod","accessor","enumMember"]}]`)},
			{Code: "\n        const obj = {\n          Foo: 42,\n          Bar() {\n            return 42;\n          },\n        };\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["PascalCase"],"selector":"property"},{"format":["PascalCase"],"selector":"method"}]`)},
			{Code: "\n        const obj = {\n          Bar() {\n            return 42;\n          },\n          async async_bar() {\n            return 42;\n          },\n        };\n        class foo {\n          public Bar() {\n            return 42;\n          }\n          public async async_bar() {\n            return 42;\n          }\n        }\n        abstract class foo {\n          public Bar() {\n            return 42;\n          }\n          public async async_bar() {\n            return 42;\n          }\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["async"],"selector":["method","objectLiteralMethod"]},{"format":["PascalCase"],"selector":"method"}]`)},
			{Code: "\n        const async_bar1 = async () => {};\n        async function async_bar2() {}\n        const async_bar3 = async function async_bar4() {};\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["PascalCase"],"selector":"method"},{"format":["snake_case"],"modifiers":["async"],"selector":["variable"]}]`)},
			{Code: "\n        class foo extends bar {\n          public someAttribute = 1;\n          public override some_attribute_override = 1;\n          public someMethod() {\n            return 42;\n          }\n          public override some_method_override2() {\n            return 42;\n          }\n        }\n        abstract class foo extends bar {\n          public abstract someAttribute: string;\n          public abstract override some_attribute_override: string;\n          public abstract someMethod(): string;\n          public abstract override some_method_override2(): string;\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["override"],"selector":["memberLike"]}]`)},
			{Code: "\n        class foo {\n          private someAttribute = 1;\n          #some_attribute = 1;\n\n          private someMethod() {}\n          #some_method() {}\n        }\n      ", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["#private"],"selector":["memberLike"]}]`)},
			{Code: "import * as FooBar from 'foo_bar';", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]},{"format":["camelCase"],"modifiers":["default"],"selector":["import"]}]`)},
			{Code: "import fooBar from 'foo_bar';", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]},{"format":["camelCase"],"modifiers":["default"],"selector":["import"]}]`)},
			{Code: "import { default as fooBar } from 'foo_bar';", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]},{"format":["camelCase"],"modifiers":["default"],"selector":["import"]}]`)},
			{Code: "import { foo_bar } from 'foo_bar';", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]},{"format":["camelCase"],"modifiers":["default"],"selector":["import"]}]`)},
			{Code: "import { \"🍎\" as Foo } from 'foo_bar';", Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]}]`)},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "const x_x = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 7, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    "const x_x = 1;",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 7, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    "\n        const child_process = require('child_process');\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"filter":{"match":true,"regex":"child_process"},"format":["camelCase"],"selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 15, EndLine: 2, EndColumn: 28},
				},
			},
			{
				Code:    "\n        declare const any_camelCase01: any;\n        declare const any_camelCase02: any | null;\n        declare const any_camelCase03: any | null | undefined;\n        declare const string_camelCase01: string;\n        declare const string_camelCase02: string | null;\n        declare const string_camelCase03: string | null | undefined;\n        declare const string_camelCase04: 'a' | null | undefined;\n        declare const string_camelCase05: string | 'a' | null | undefined;\n        declare const number_camelCase06: number;\n        declare const number_camelCase07: number | null;\n        declare const number_camelCase08: number | null | undefined;\n        declare const number_camelCase09: 1 | null | undefined;\n        declare const number_camelCase10: number | 2 | null | undefined;\n        declare const boolean_camelCase11: boolean;\n        declare const boolean_camelCase12: boolean | null;\n        declare const boolean_camelCase13: boolean | null | undefined;\n        declare const boolean_camelCase14: true | null | undefined;\n        declare const boolean_camelCase15: false | null | undefined;\n        declare const boolean_camelCase16: true | false | null | undefined;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["UPPER_CASE"],"modifiers":["const"],"prefix":["any_"],"selector":"variable"},{"format":["snake_case"],"prefix":["string_"],"selector":"variable","types":["string"]},{"format":["snake_case"],"prefix":["number_"],"selector":"variable","types":["number"]},{"format":["snake_case"],"prefix":["boolean_"],"selector":"variable","types":["boolean"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 23, EndLine: 2, EndColumn: 43},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 3, Column: 23, EndLine: 3, EndColumn: 50},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 4, Column: 23, EndLine: 4, EndColumn: 62},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 5, Column: 23, EndLine: 5, EndColumn: 49},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 6, Column: 23, EndLine: 6, EndColumn: 56},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 7, Column: 23, EndLine: 7, EndColumn: 68},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 8, Column: 23, EndLine: 8, EndColumn: 65},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 9, Column: 23, EndLine: 9, EndColumn: 74},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 10, Column: 23, EndLine: 10, EndColumn: 49},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 11, Column: 23, EndLine: 11, EndColumn: 56},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 12, Column: 23, EndLine: 12, EndColumn: 68},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 13, Column: 23, EndLine: 13, EndColumn: 63},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 14, Column: 23, EndLine: 14, EndColumn: 72},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 15, Column: 23, EndLine: 15, EndColumn: 51},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 16, Column: 23, EndLine: 16, EndColumn: 58},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 17, Column: 23, EndLine: 17, EndColumn: 70},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 18, Column: 23, EndLine: 18, EndColumn: 67},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 19, Column: 23, EndLine: 19, EndColumn: 68},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 20, Column: 23, EndLine: 20, EndColumn: 75},
				},
			},
			{
				Code:    "\n        declare const function_camelCase1: () => void;\n        declare const function_camelCase2: (() => void) | null;\n        declare const function_camelCase3: (() => void) | null | undefined;\n        declare const function_camelCase4:\n          | (() => void)\n          | (() => string)\n          | null\n          | undefined;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"prefix":["function_"],"selector":"variable","types":["function"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 23, EndLine: 2, EndColumn: 54},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 3, Column: 23, EndLine: 3, EndColumn: 63},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 4, Column: 23, EndLine: 4, EndColumn: 75},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 5, Column: 23, EndLine: 9, EndColumn: 22},
				},
			},
			{
				Code:    "\n        declare const array_camelCase1: Array<number>;\n        declare const array_camelCase2: ReadonlyArray<number> | null;\n        declare const array_camelCase3: number[] | null | undefined;\n        declare const array_camelCase4: readonly number[] | null | undefined;\n        declare const array_camelCase5:\n          | number[]\n          | (number | string)[]\n          | null\n          | undefined;\n        declare const array_camelCase6: [] | null | undefined;\n        declare const array_camelCase7: [number] | null | undefined;\n        declare const array_camelCase8:\n          | readonly number[]\n          | Array<string>\n          | [boolean]\n          | null\n          | undefined;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"prefix":["array_"],"selector":"variable","types":["array"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 23, EndLine: 2, EndColumn: 54},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 3, Column: 23, EndLine: 3, EndColumn: 69},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 4, Column: 23, EndLine: 4, EndColumn: 68},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 5, Column: 23, EndLine: 5, EndColumn: 77},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 6, Column: 23, EndLine: 10, EndColumn: 22},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 11, Column: 23, EndLine: 11, EndColumn: 62},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 12, Column: 23, EndLine: 12, EndColumn: 68},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 13, Column: 23, EndLine: 18, EndColumn: 22},
				},
			},
			{
				Code:    "\n        let unused_foo = 'a';\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^unused_\\w"},"format":["snake_case"],"leadingUnderscore":"allow","selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 13, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    "\n        const _unused_foo = 1;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^unused_\\w"},"format":["snake_case"],"leadingUnderscore":"allow","selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 15, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code:    "\n        interface IFoo {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^I[A-Z]"},"format":["PascalCase"],"selector":"typeLike"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 19, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    "\n        class IBar {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^I[A-Z]"},"format":["PascalCase"],"selector":"typeLike"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 15, EndLine: 2, EndColumn: 19},
				},
			},
			{
				Code:    "\n        function fooBar() {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":true,"regex":"function"},"format":["camelCase"],"leadingUnderscore":"allow","selector":"function"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 18, EndLine: 2, EndColumn: 24},
				},
			},
			{
				Code:    "\n        let unused_foo = 'a';\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"leadingUnderscore":"allow","selector":["variable","function"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 13, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    "\n        const _unused_foo = 1;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"leadingUnderscore":"allow","selector":["variable","function"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 15, EndLine: 2, EndColumn: 26},
				},
			},
			{
				Code:    "\n        function foo_bar() {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"leadingUnderscore":"allow","selector":["variable","function"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 18, EndLine: 2, EndColumn: 25},
				},
			},
			{
				Code:    "\n        interface IFoo {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"custom":{"match":false,"regex":"^I[A-Z]"},"format":["PascalCase"],"selector":["class","interface"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 19, EndLine: 2, EndColumn: 23},
				},
			},
			{
				Code:    "\n        class IBar {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"leadingUnderscore":"allow","selector":["variable","function"]},{"custom":{"match":false,"regex":"^I[A-Z]"},"format":["PascalCase"],"selector":["class","interface"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "satisfyCustom", Line: 2, Column: 15, EndLine: 2, EndColumn: 19},
				},
			},
			{
				Code:    "\n        const foo = {\n          'Property Name': 'asdf',\n        };\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"filter":{"match":false,"regex":"-"},"format":["strictCamelCase"],"selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 11, EndLine: 3, EndColumn: 26},
				},
			},
			{
				Code:    "\n        const myfoo_bar = 'abcs';\n        function fun(myfoo: string) {}\n        class foo {\n          Myfoo: string;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"prefix":["my","My"],"selector":["variable","property","parameter"],"types":["string"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 15, EndLine: 2, EndColumn: 24},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 3, Column: 22, EndLine: 3, EndColumn: 35},
					{MessageId: "doesNotMatchFormatTrimmed", Line: 5, Column: 11, EndLine: 5, EndColumn: 16},
				},
			},
			{
				Code:    "\n        class foo {\n          private readonly fooBar: boolean;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"modifiers":["private","readonly"],"selector":["property","accessor"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 28, EndLine: 3, EndColumn: 34},
				},
			},
			{
				Code:    "\n        function my_foo_bar() {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"prefix":["my","My"],"selector":["variable","function"],"types":["string"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormatTrimmed", Line: 2, Column: 18, EndLine: 2, EndColumn: 28},
				},
			},
			{
				Code:    "\n        class SomeClass {\n          static otherConstant = 'hello';\n        }\n\n        export const { otherConstant } = SomeClass;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"property"},{"format":["camelCase"],"selector":"variable"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 18, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code:    "\n        declare class Foo {\n          Bar(Baz: string): void;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"parameter"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 15, EndLine: 3, EndColumn: 26},
				},
			},
			{
				Code:    "\n        export const PascalCaseVar = 1;\n        export enum PascalCaseEnum {}\n        export class PascalCaseClass {}\n        export function PascalCaseFunction() {}\n        export interface PascalCaseInterface {}\n        export type PascalCaseType = {};\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":["camelCase"],"modifiers":["exported"],"selector":"variable"},{"format":["camelCase"],"modifiers":["exported"],"selector":"function"},{"format":["camelCase"],"modifiers":["exported"],"selector":"class"},{"format":["camelCase"],"modifiers":["exported"],"selector":"interface"},{"format":["camelCase"],"modifiers":["exported"],"selector":"typeAlias"},{"format":["camelCase"],"modifiers":["exported"],"selector":"enum"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 22, EndLine: 2, EndColumn: 35},
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 21, EndLine: 3, EndColumn: 35},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 22, EndLine: 4, EndColumn: 37},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 25, EndLine: 5, EndColumn: 43},
					{MessageId: "doesNotMatchFormat", Line: 6, Column: 26, EndLine: 6, EndColumn: 45},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 21, EndLine: 7, EndColumn: 35},
				},
			},
			{
				Code:    "\n        const PascalCaseVar = 1;\n        enum PascalCaseEnum {}\n        class PascalCaseClass {}\n        function PascalCaseFunction() {}\n        interface PascalCaseInterface {}\n        type PascalCaseType = {};\n        export {\n          PascalCaseVar,\n          PascalCaseEnum,\n          PascalCaseClass,\n          PascalCaseFunction,\n          PascalCaseInterface,\n          PascalCaseType,\n        };\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":["camelCase"],"modifiers":["exported"],"selector":"variable"},{"format":["camelCase"],"modifiers":["exported"],"selector":"function"},{"format":["camelCase"],"modifiers":["exported"],"selector":"class"},{"format":["camelCase"],"modifiers":["exported"],"selector":"interface"},{"format":["camelCase"],"modifiers":["exported"],"selector":"typeAlias"},{"format":["camelCase"],"modifiers":["exported"],"selector":"enum"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 15, EndLine: 2, EndColumn: 28},
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 14, EndLine: 3, EndColumn: 28},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 15, EndLine: 4, EndColumn: 30},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 18, EndLine: 5, EndColumn: 36},
					{MessageId: "doesNotMatchFormat", Line: 6, Column: 19, EndLine: 6, EndColumn: 38},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 14, EndLine: 7, EndColumn: 28},
				},
			},
			{
				Code:    "\n        const PascalCaseVar = 1;\n        function PascalCaseFunction() {}\n        declare function PascalCaseDeclaredFunction();\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":["camelCase"],"modifiers":["global"],"selector":"variable"},{"format":["camelCase"],"modifiers":["global"],"selector":"function"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 15, EndLine: 2, EndColumn: 28},
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 18, EndLine: 3, EndColumn: 36},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 26, EndLine: 4, EndColumn: 52},
				},
			},
			{
				Code:    "\n        const { some_name1 } = {};\n        const { some_name2 = 2 } = {};\n        const { ignored: IgnoredDueToModifiers1 } = {};\n        const { ignored: IgnoredDueToModifiers2 = 3 } = {};\n        const IgnoredDueToModifiers3 = 1;\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["destructured"],"selector":"variable"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 17, EndLine: 2, EndColumn: 27},
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 17, EndLine: 3, EndColumn: 27},
				},
			},
			{
				Code:    "\n        export function Foo(\n          { aName },\n          { anotherName = 1 },\n          { ignored: IgnoredDueToModifiers1 },\n          { ignored: IgnoredDueToModifiers1 = 2 },\n          IgnoredDueToModifiers2,\n        ) {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["destructured"],"selector":"parameter"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 13, EndLine: 3, EndColumn: 18},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 13, EndLine: 4, EndColumn: 24},
				},
			},
			{
				Code:    "\n        class Ignored {\n          private static abstract readonly some_name;\n          IgnoredDueToModifiers = 1;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["static","readonly"],"selector":"classProperty"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 44, EndLine: 3, EndColumn: 53},
				},
			},
			{
				Code:    "\n        class Ignored {\n          constructor(\n            private readonly some_name,\n            IgnoredDueToModifiers,\n          ) {}\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["readonly"],"selector":"parameterProperty"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 30, EndLine: 4, EndColumn: 39},
				},
			},
			{
				Code:    "\n        class Ignored {\n          private static some_name() {}\n          IgnoredDueToModifiers() {}\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["static"],"selector":"classMethod"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 26, EndLine: 3, EndColumn: 35},
				},
			},
			{
				Code:    "\n        class Ignored {\n          private static get some_name() {}\n          get IgnoredDueToModifiers() {}\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["private","static"],"selector":"accessor"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 30, EndLine: 3, EndColumn: 39},
				},
			},
			{
				Code:    "\n        abstract class some_name {}\n        class IgnoredDueToModifier {}\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["UPPER_CASE"],"modifiers":["abstract"],"selector":"class"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 24, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "\n        const UnusedVar = 1;\n        function UnusedFunc(\n          // this line is intentionally broken out\n          UnusedParam: string,\n        ) {}\n        class UnusedClass {}\n        interface UnusedInterface {}\n        type UnusedType<\n          // this line is intentionally broken out\n          UnusedTypeParam,\n        > = {};\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":"default"},{"format":["snake_case"],"modifiers":["unused"],"selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 2, Column: 15, EndLine: 2, EndColumn: 24},
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 18, EndLine: 3, EndColumn: 28},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 11, EndLine: 5, EndColumn: 30},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 15, EndLine: 7, EndColumn: 26},
					{MessageId: "doesNotMatchFormat", Line: 8, Column: 19, EndLine: 8, EndColumn: 34},
					{MessageId: "doesNotMatchFormat", Line: 9, Column: 14, EndLine: 9, EndColumn: 24},
					{MessageId: "doesNotMatchFormat", Line: 11, Column: 11, EndLine: 11, EndColumn: 26},
				},
			},
			{
				Code:    "\n        const ignored1 = {\n          'a a': 1,\n          'b b'() {},\n          get 'c c'() {\n            return 1;\n          },\n          set 'd d'(value: string) {},\n        };\n        class ignored2 {\n          'a a' = 1;\n          'b b'() {}\n          get 'c c'() {\n            return 1;\n          }\n          set 'd d'(value: string) {}\n        }\n        interface ignored3 {\n          'a a': 1;\n          'b b'(): void;\n        }\n        type ignored4 = {\n          'a a': 1;\n          'b b'(): void;\n        };\n        enum ignored5 {\n          'a a',\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["snake_case"],"selector":"default"},{"format":["PascalCase"],"modifiers":["requiresQuotes"],"selector":"default"}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 11, EndLine: 3, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 11, EndLine: 4, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 15, EndLine: 5, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 8, Column: 15, EndLine: 8, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 11, Column: 11, EndLine: 11, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 12, Column: 11, EndLine: 12, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 13, Column: 15, EndLine: 13, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 16, Column: 15, EndLine: 16, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 19, Column: 11, EndLine: 19, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 20, Column: 11, EndLine: 20, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 23, Column: 11, EndLine: 23, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 24, Column: 11, EndLine: 24, EndColumn: 16},
					{MessageId: "doesNotMatchFormat", Line: 27, Column: 11, EndLine: 27, EndColumn: 16},
				},
			},
			{
				Code: "\n        type Foo = {\n          'foo     Bar': string;\n          '': string;\n          '0': string;\n          'foo': string;\n          'foo-bar': string;\n          '#foo-bar': string;\n        };\n\n        interface Bar {\n          'boo-----foo': string;\n        }\n      ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 3, Column: 11, EndLine: 3, EndColumn: 24},
					{MessageId: "doesNotMatchFormat", Line: 4, Column: 11, EndLine: 4, EndColumn: 13},
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 11, EndLine: 5, EndColumn: 14},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 11, EndLine: 7, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 8, Column: 11, EndLine: 8, EndColumn: 21},
					{MessageId: "doesNotMatchFormat", Line: 12, Column: 11, EndLine: 12, EndColumn: 24},
				},
			},
			{
				Code:    "\n        class foo {\n          public Bar() {\n            return 42;\n          }\n          public async async_bar() {\n            return 42;\n          }\n          // ❌ error\n          public async asyncBar() {\n            return 42;\n          }\n          // ❌ error\n          public AsyncBar2 = async () => {\n            return 42;\n          };\n          // ❌ error\n          public AsyncBar3 = async function () {\n            return 42;\n          };\n        }\n        abstract class foo {\n          public abstract Bar(): number;\n          public abstract async async_bar(): number;\n          // ❌ error\n          public abstract async ASYNC_BAR(): number;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["PascalCase"],"selector":"method"},{"format":["snake_case"],"modifiers":["async"],"selector":["method","objectLiteralMethod"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 10, Column: 24, EndLine: 10, EndColumn: 32},
					{MessageId: "doesNotMatchFormat", Line: 14, Column: 18, EndLine: 14, EndColumn: 27},
					{MessageId: "doesNotMatchFormat", Line: 18, Column: 18, EndLine: 18, EndColumn: 27},
					{MessageId: "doesNotMatchFormat", Line: 26, Column: 33, EndLine: 26, EndColumn: 42},
				},
			},
			{
				Code:    "\n        const obj = {\n          Bar() {\n            return 42;\n          },\n          async async_bar() {\n            return 42;\n          },\n          // ❌ error\n          async AsyncBar() {\n            return 42;\n          },\n          // ❌ error\n          AsyncBar2: async () => {\n            return 42;\n          },\n          // ❌ error\n          AsyncBar3: async function () {\n            return 42;\n          },\n        };\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["PascalCase"],"selector":"method"},{"format":["snake_case"],"modifiers":["async"],"selector":["method","objectLiteralMethod"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 10, Column: 17, EndLine: 10, EndColumn: 25},
					{MessageId: "doesNotMatchFormat", Line: 14, Column: 11, EndLine: 14, EndColumn: 20},
					{MessageId: "doesNotMatchFormat", Line: 18, Column: 11, EndLine: 18, EndColumn: 20},
				},
			},
			{
				Code:    "\n        const syncbar1 = () => {};\n        function syncBar2() {}\n        const syncBar3 = function syncBar4() {};\n\n        // ❌ error\n        const AsyncBar1 = async () => {};\n        const async_bar1 = async () => {};\n        const async_bar3 = async function async_bar4() {};\n        async function async_bar2() {}\n        // ❌ error\n        const asyncBar5 = async function async_bar6() {};\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"variableLike"},{"format":["snake_case"],"modifiers":["async"],"selector":["variableLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 15, EndLine: 7, EndColumn: 24},
					{MessageId: "doesNotMatchFormat", Line: 12, Column: 15, EndLine: 12, EndColumn: 24},
				},
			},
			{
				Code:    "\n        const syncbar1 = () => {};\n        function syncBar2() {}\n        const syncBar3 = function syncBar4() {};\n\n        const async_bar1 = async () => {};\n        // ❌ error\n        async function asyncBar2() {}\n        const async_bar3 = async function async_bar4() {};\n        async function async_bar2() {}\n        // ❌ error\n        const async_bar3 = async function ASYNC_BAR4() {};\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"variableLike"},{"format":["snake_case"],"modifiers":["async"],"selector":["variableLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 8, Column: 24, EndLine: 8, EndColumn: 33},
					{MessageId: "doesNotMatchFormat", Line: 12, Column: 43, EndLine: 12, EndColumn: 53},
				},
			},
			{
				Code:    "\n        class foo extends bar {\n          public someAttribute = 1;\n          public override some_attribute_override = 1;\n          // ❌ error\n          public override someAttributeOverride = 1;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["override"],"selector":["memberLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 6, Column: 27, EndLine: 6, EndColumn: 48},
				},
			},
			{
				Code:    "\n        class foo extends bar {\n          public override some_method_override() {\n            return 42;\n          }\n          // ❌ error\n          public override someMethodOverride() {\n            return 42;\n          }\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["override"],"selector":["memberLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 27, EndLine: 7, EndColumn: 45},
				},
			},
			{
				Code:    "\n        class foo extends bar {\n          public get someGetter(): string;\n          public override get some_getter_override(): string;\n          // ❌ error\n          public override get someGetterOverride(): string;\n          public set someSetter(val: string);\n          public override set some_setter_override(val: string);\n          // ❌ error\n          public override set someSetterOverride(val: string);\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["override"],"selector":["memberLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 6, Column: 31, EndLine: 6, EndColumn: 49},
					{MessageId: "doesNotMatchFormat", Line: 10, Column: 31, EndLine: 10, EndColumn: 49},
				},
			},
			{
				Code:    "\n        class foo {\n          private firstPrivateField = 1;\n          // ❌ error\n          private first_private_field = 1;\n          // ❌ error\n          #secondPrivateField = 1;\n          #second_private_field = 1;\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["#private"],"selector":["memberLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 19, EndLine: 5, EndColumn: 38},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 11, EndLine: 7, EndColumn: 30},
				},
			},
			{
				Code:    "\n        class foo {\n          private firstPrivateMethod() {}\n          // ❌ error\n          private first_private_method() {}\n          // ❌ error\n          #secondPrivateMethod() {}\n          #second_private_method() {}\n        }\n      ",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":"memberLike"},{"format":["snake_case"],"modifiers":["#private"],"selector":["memberLike"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 5, Column: 19, EndLine: 5, EndColumn: 39},
					{MessageId: "doesNotMatchFormat", Line: 7, Column: 11, EndLine: 7, EndColumn: 31},
				},
			},
			{
				Code:    "import * as fooBar from 'foo_bar';",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":["import"]},{"format":["PascalCase"],"modifiers":["namespace"],"selector":["import"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "import FooBar from 'foo_bar';",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":["import"]},{"format":["PascalCase"],"modifiers":["namespace"],"selector":["import"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "import { default as foo_bar } from 'foo_bar';",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["camelCase"],"selector":["import"]},{"format":["PascalCase"],"modifiers":["namespace"],"selector":["import"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "import { \"🍎\" as foo } from 'foo_bar';",
				Options: rule_tester.OptionsFromJSON[NamingConventionOptions](`[{"format":["PascalCase"],"selector":["import"]}]`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "doesNotMatchFormat", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
				},
			},
		},
	)
}
