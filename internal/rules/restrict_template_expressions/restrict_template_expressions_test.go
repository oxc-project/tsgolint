package restrict_template_expressions

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestRestrictTemplateExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RestrictTemplateExpressionsRule, []rule_tester.ValidTestCase{
		{Code: `
      const msg = ` + "`" + `arg = ${'foo'}` + "`" + `;
    `},
		{Code: `
      const arg = 'foo';
      const msg = ` + "`" + `arg = ${arg}` + "`" + `;
    `},
		{Code: `
      const arg = 'foo';
      const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
    `},
		{Code: `
      function test<T extends string>(arg: T) {
        return ` + "`" + `arg = ${arg}` + "`" + `;
      }
    `},
		{Code: `
      function test<T extends string & { _kind: 'MyBrandedString' }>(arg: T) {
        return ` + "`" + `arg = ${arg}` + "`" + `;
      }
    `},
		{Code: `
      tag` + "`" + `arg = ${null}` + "`" + `;
    `},
		{Code: `
      const arg = {};
      tag` + "`" + `arg = ${arg}` + "`" + `;
    `},
		{
			Code: `
        const arg = 123;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        const arg = 123;
        const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        const arg = 123n;
        const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        function test<T extends number>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        function test<T extends number & { _kind: 'MyBrandedNumber' }>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        function test<T extends bigint>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        function test<T extends string | number>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": true}`),
		},
		{
			Code: `
        const arg = true;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true}`),
		},
		{
			Code: `
        const arg = true;
        const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true}`),
		},
		{
			Code: `
        function test<T extends boolean>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true}`),
		},
		{
			Code: `
        function test<T extends string | boolean>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true}`),
		},
		{
			Code: `
        const arg = [];
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        const arg = [];
        const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        function test<T extends string[]>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        declare const arg: [number, string];
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        const arg = [1, 'a'] as const;
        const msg = ` + "`" + `arg = ${arg || 'default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        function test<T extends [string, string]>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true}`),
		},
		{
			Code: `
        declare const arg: [number | undefined, string];
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true, "allowNullish": true}`),
		},
		{
			Code: `
        const arg: any = 123;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": true}`),
		},
		{
			Code: `
        const arg: any = undefined;
        const msg = ` + "`" + `arg = ${arg || 'some-default'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": true}`),
		},
		{
			Code: `
        const user = JSON.parse('{ "name": "foo" }');
        const msg = ` + "`" + `arg = ${user.name}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": true}`),
		},
		{
			Code: `
        const user = JSON.parse('{ "name": "foo" }');
        const msg = ` + "`" + `arg = ${user.name || 'the user with no name'}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": true}`),
		},
		{
			Code: `
        const arg = null;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNullish": true}`),
		},
		{
			Code: `
        declare const arg: string | null | undefined;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNullish": true}`),
		},
		{
			Code: `
        function test<T extends null | undefined>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNullish": true}`),
		},
		{
			Code: `
        function test<T extends string | null>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNullish": true}`),
		},
		{
			Code: `
        const arg = new RegExp('foo');
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": true}`),
		},
		{
			Code: `
        const arg = /foo/;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": true}`),
		},
		{
			Code: `
        declare const arg: string | RegExp;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": true}`),
		},
		{
			Code: `
        function test<T extends RegExp>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": true}`),
		},
		{
			Code: `
        function test<T extends string | RegExp>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": true}`),
		},
		{
			Code: `
        declare const value: never;
        const stringy = ` + "`" + `${value}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNever": true}`),
		},
		{
			Code: `
        const arg = 'hello';
        const msg = typeof arg === 'string' ? arg : ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNever": true}`),
		},
		{
			Code: `
        function test(arg: 'one' | 'two') {
          switch (arg) {
            case 'one':
              return 1;
            case 'two':
              return 2;
            default:
              throw new Error(` + "`" + `Unrecognized arg: ${arg}` + "`" + `);
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNever": true}`),
		},
		{
			Code: `
        // more variants may be added to Foo in the future
        type Foo = { type: 'a'; value: number };

        function checkFoosAreMatching(foo1: Foo, foo2: Foo) {
          if (foo1.type !== foo2.type) {
            // since Foo currently only has one variant, this code is never run, and ` + "`" + `foo1.type` + "`" + ` has type ` + "`" + `never` + "`" + `.
            throw new Error(` + "`" + `expected ${foo1.type}, found ${foo2.type}` + "`" + `);
          }
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNever": true}`),
		},
		{
			Code: `
        type All = string | number | boolean | null | undefined | RegExp | never;
        function test<T extends All>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true, "allowNever": true, "allowNullish": true, "allowNumber": true, "allowRegExp": true}`),
		},
		{
			Code:    "const msg = `arg = ${Promise.resolve()}`;",
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allow": [{"from": "lib", "name": ["Promise"]}]}`),
		},
		{Code: "const msg = `arg = ${new Error()}`;"},
		{Code: "const msg = `arg = ${false}`;"},
		{Code: "const msg = `arg = ${null}`;"},
		{Code: "const msg = `arg = ${undefined}`;"},
		{Code: "const msg = `arg = ${123}`;"},
		{Code: "const msg = `arg = ${'abc'}`;"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        const msg = ` + "`" + `arg = ${123}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      2,
					Column:    30,
				},
			},
		},
		{
			Code: `
        const msg = ` + "`" + `arg = ${false}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      2,
					Column:    30,
				},
			},
		},
		{
			Code: `
        const msg = ` + "`" + `arg = ${null}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNullish": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      2,
					Column:    30,
				},
			},
		},
		{
			Code: `
        declare const arg: number[];
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        const msg = ` + "`" + `arg = ${[, 2]}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true, "allowNullish": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      2,
					Column:    30,
				},
			},
		},
		{
			Code: "const msg = `arg = ${Promise.resolve()}`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
				},
			},
		},
		{
			Code:    "const msg = `arg = ${new Error()}`;",
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allow": []}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
				},
			},
		},
		{
			Code: `
        declare const arg: [number | undefined, string];
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowArray": true, "allowNullish": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        declare const arg: number;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNumber": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        declare const arg: boolean;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        const arg = {};
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true, "allowNullish": true, "allowNumber": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        declare const arg: { a: string } & { b: string };
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        function test<T extends {}>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowBoolean": true, "allowNullish": true, "allowNumber": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    27,
				},
			},
		},
		{
			Code: `
        function test<TWithNoConstraint>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": false, "allowBoolean": true, "allowNullish": true, "allowNumber": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    27,
				},
			},
		},
		{
			Code: `
        function test(arg: any) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": false, "allowBoolean": true, "allowNullish": true, "allowNumber": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    27,
				},
			},
		},
		{
			Code: `
        const arg = new RegExp('foo');
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        const arg = /foo/;
        const msg = ` + "`" + `arg = ${arg}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowRegExp": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    30,
				},
			},
		},
		{
			Code: `
        declare const value: never;
        const stringy = ` + "`" + `${value}` + "`" + `;
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowNever": false}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    28,
				},
			},
		},
		{
			Code: `
        function test<T extends any>(arg: T) {
          return ` + "`" + `arg = ${arg}` + "`" + `;
        }
      `,
			Options: rule_tester.OptionsFromJSON[RestrictTemplateExpressionsOptions](`{"allowAny": true}`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidType",
					Line:      3,
					Column:    27,
				},
			},
		},
	})
}
