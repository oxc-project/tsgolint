package prefer_destructuring

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferDestructuringRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferDestructuringRule, []rule_tester.ValidTestCase{
		// type annotated
		{Code: `
      declare const object: { foo: string };
      var foo: string = object.foo;
    `},
		{Code: `
      declare const array: number[];
      const bar: number = array[0];
    `},
		// enforceForDeclarationWithTypeAnnotation: true
		{Code: `
        declare const object: { foo: string };
        var { foo } = object;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const object: { foo: string };
        var { foo }: { foo: number } = object;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const array: number[];
        var [foo] = array;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const array: number[];
        var [foo]: [foo: number] = array;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const object: { bar: string };
        var foo: unknown = object.bar;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const object: { foo: string };
        var { foo: bar } = object;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare const object: { foo: boolean };
        var { foo: bar }: { foo: boolean } = object;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        declare class Foo {
          foo: string;
        }

        class Bar extends Foo {
          static foo() {
            var foo: any = super.foo;
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},

		// numeric property for iterable / non-iterable
		{Code: `
      let x: { 0: unknown };
      let y = x[0];
    `},
		{Code: `
      let x: { 0: unknown };
      y = x[0];
    `},
		{Code: `
      let x: unknown;
      let y = x[0];
    `},
		{Code: `
      let x: unknown;
      y = x[0];
    `},
		{Code: `
      let x: { 0: unknown } | unknown[];
      let y = x[0];
    `},
		{Code: `
      let x: { 0: unknown } | unknown[];
      y = x[0];
    `},
		{Code: `
      let x: { 0: unknown } & (() => void);
      let y = x[0];
    `},
		{Code: `
      let x: { 0: unknown } & (() => void);
      y = x[0];
    `},
		{Code: `
      let x: Record<number, unknown>;
      let y = x[0];
    `},
		{Code: `
      let x: Record<number, unknown>;
      y = x[0];
    `},
		{Code: `
        let x: { 0: unknown };
        let { 0: y } = x;
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: { 0: unknown };
        ({ 0: y } = x);
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: { 0: unknown };
        let y = x[0];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: { 0: unknown };
        y = x[0];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: { 0: unknown };
        let y = x[0];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"AssignmentExpression":{"array":true,"object":true},"VariableDeclarator":{"array":true,"object":false},"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: { 0: unknown };
        y = x[0];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"AssignmentExpression":{"array":true,"object":false},"VariableDeclarator":{"array":true,"object":true},"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: Record<number, unknown>;
        let i: number = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: Record<number, unknown>;
        let i: 0 = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: Record<number, unknown>;
        let i: 0 | 1 | 2 = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: unknown[];
        let i: number = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: unknown[];
        let i: 0 = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: unknown[];
        let i: 0 | 1 | 2 = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":false,"enforceForRenamedProperties":true}`)},
		{Code: `
        let x: unknown[];
        let i: number = 0;
        y = x[i];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":false}`)},
		{Code: `
        let x: { 0: unknown };
        y += x[0];
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        class Bar {
          public [0]: unknown;
        }
        class Foo extends Bar {
          static foo() {
            let y = super[0];
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`)},
		{Code: `
        class Bar {
          public [0]: unknown;
        }
        class Foo extends Bar {
          static foo() {
            y = super[0];
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`)},

		// already destructured
		{Code: `
      let xs: unknown[] = [1];
      let [x] = xs;
    `},
		{Code: `
      const obj: { x: unknown } = { x: 1 };
      const { x } = obj;
    `},
		{Code: `
      var obj: { x: unknown } = { x: 1 };
      var { x: y } = obj;
    `},
		{Code: `
      let obj: { x: unknown } = { x: 1 };
      let key: 'x' = 'x';
      let { [key]: foo } = obj;
    `},
		{Code: `
      const obj: { x: unknown } = { x: 1 };
      let x: unknown;
      ({ x } = obj);
    `},

		// valid unless enforceForRenamedProperties is true
		{Code: `
      let obj: { x: unknown } = { x: 1 };
      let y = obj.x;
    `},
		{Code: `
      var obj: { x: unknown } = { x: 1 };
      var y: unknown;
      y = obj.x;
    `},
		{Code: `
      const obj: { x: unknown } = { x: 1 };
      const y = obj['x'];
    `},
		{Code: `
      let obj: Record<string, unknown> = {};
      let key = 'abc';
      var y = obj[key];
    `},

		// shorthand operators shouldn't be reported
		{Code: `
      let obj: { x: number } = { x: 1 };
      let x = 10;
      x += obj.x;
    `},
		{Code: `
      let obj: { x: boolean } = { x: false };
      let x = true;
      x ||= obj.x;
    `},
		{Code: `
      const xs: number[] = [1];
      let x = 3;
      x *= xs[0];
    `},

		// optional chaining shouldn't be reported
		{Code: `
      let xs: unknown[] | undefined;
      let x = xs?.[0];
    `},
		{Code: `
      let obj: Record<string, unknown> | undefined;
      let x = obj?.x;
    `},

		// private identifiers
		{Code: `
      class C {
        #foo: string;

        method() {
          const foo: unknown = this.#foo;
        }
      }
    `},
		{Code: `
      class C {
        #foo: string;

        method() {
          let foo: unknown;
          foo = this.#foo;
        }
      }
    `},
		{Code: `
        class C {
          #foo: string;

          method() {
            const bar: unknown = this.#foo;
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        class C {
          #foo: string;

          method(another: C) {
            let bar: unknown;
            bar: unknown = another.#foo;
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
		{Code: `
        class C {
          #foo: string;

          method() {
            const foo: unknown = this.#foo;
          }
        }
      `, Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForDeclarationWithTypeAnnotation":true}`)},
	}, []rule_tester.InvalidTestCase{
		// enforceForDeclarationWithTypeAnnotation: true
		{
			Code:    `var foo: string = object.foo;`,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code:    `var foo: string = array[0];`,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"enforceForDeclarationWithTypeAnnotation":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code:    `var foo: unknown = object.bar;`,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForDeclarationWithTypeAnnotation":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},

		// numeric property for iterable / non-iterable
		{
			Code: `
        let x: { [Symbol.iterator]: unknown };
        let y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { [Symbol.iterator]: unknown };
        y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: [1, 2, 3];
        let y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: [1, 2, 3];
        y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        function* it() {
          yield 1;
        }
        let y = it()[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        function* it() {
          yield 1;
        }
        y = it()[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: any;
        let y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: any;
        y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: string[] | { [Symbol.iterator]: unknown };
        let y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: string[] | { [Symbol.iterator]: unknown };
        y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: object & unknown[];
        let y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: object & unknown[];
        y = x[0];
      `,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: string };
        let y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: string };
        y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: string };
        let y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"AssignmentExpression":{"array":false,"object":false},"VariableDeclarator":{"array":false,"object":true},"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: string };
        y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"AssignmentExpression":{"array":false,"object":true},"VariableDeclarator":{"array":false,"object":false},"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: Record<number, unknown>;
        let i: number = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: Record<number, unknown>;
        let i: 0 = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: Record<number, unknown>;
        let i: 0 | 1 | 2 = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: unknown[];
        let i: number = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: unknown[];
        let i: 0 = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: unknown[];
        let i: 0 | 1 | 2 = 0;
        y = x[i];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"array":true,"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: unknown } | unknown[];
        let y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let x: { 0: unknown } | unknown[];
        y = x[0];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},

		// auto fixes
		{
			Code: `
        let obj = { foo: 'bar' };
        const foo = obj.foo;
      `,
			Output: []string{`
        let obj = { foo: 'bar' };
        const {foo} = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let obj = { foo: 'bar' };
        var x: null = null;
        const foo = (x, obj).foo;
      `,
			Output: []string{`
        let obj = { foo: 'bar' };
        var x: null = null;
        const {foo} = (x, obj);
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code:   `const call = (() => null).call;`,
			Output: []string{`const {call} = () => null;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        const obj = { foo: 'bar' };
        let a: any;
        var foo = (a = obj).foo;
      `,
			Output: []string{`
        const obj = { foo: 'bar' };
        let a: any;
        var {foo} = a = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        const obj = { asdf: { qwer: null } };
        const qwer = obj.asdf.qwer;
      `,
			Output: []string{`
        const obj = { asdf: { qwer: null } };
        const {qwer} = obj.asdf;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        const obj = { foo: 100 };
        const /* comment */ foo = obj.foo;
      `,
			Output: []string{`
        const obj = { foo: 100 };
        const /* comment */ {foo} = obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},

		// enforceForRenamedProperties: true
		{
			Code: `
        let obj = { foo: 'bar' };
        const x = obj.foo;
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let obj = { foo: 'bar' };
        let x: unknown;
        x = obj.foo;
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let obj: Record<string, unknown>;
        let key = 'abc';
        const x = obj[key];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
		{
			Code: `
        let obj: Record<string, unknown>;
        let key = 'abc';
        let x: unknown;
        x = obj[key];
      `,
			Options: rule_tester.OptionsFromJSON[PreferDestructuringOptions](`{"object":true,"enforceForRenamedProperties":true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferDestructuring"}},
		},
	})
}
