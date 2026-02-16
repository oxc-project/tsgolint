package no_useless_default_assignment

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUselessDefaultAssignmentRule(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUselessDefaultAssignmentRule, []rule_tester.ValidTestCase{
		{
			Code: "\n      function Bar({ foo = '' }: { foo?: string }) {\n        return foo;\n      }\n    ",
		},
		{
			Code: "\n      const { foo } = { foo: 'bar' };\n    ",
		},
		{
			Code: "\n      [1, 2, 3, undefined].map((a = 42) => a + 1);\n    ",
		},
		{
			Code: "\n      function test(a?: number) {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      const obj: { a?: string } = {};\n      const { a = 'default' } = obj;\n    ",
		},
		{
			Code: "\n      function test(a: string | undefined = 'default') {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      (a: string = 'default') => a;\n    ",
		},
		{
			Code: "\n      function test(a: string = 'default') {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      class C {\n        public test(a: string = 'default') {\n          return a;\n        }\n      }\n    ",
		},
		{
			Code: "\n      const obj: { a: string | undefined } = { a: undefined };\n      const { a = 'default' } = obj;\n    ",
		},
		{
			Code: "\n      function test(arr: number[] | undefined = []) {\n        return arr;\n      }\n    ",
		},
		{
			Code: "\n      function Bar({ nested: { foo = '' } = {} }: { nested?: { foo?: string } }) {\n        return foo;\n      }\n    ",
		},
		{
			Code: "\n      function test(a: any = 'default') {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      function test(a: unknown = 'default') {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      function test(a = 5) {\n        return a;\n      }\n    ",
		},
		{
			Code: "\n      function createValidator(): () => void {\n        return (param = 5) => {};\n      }\n    ",
		},
		{
			Code: "\n      function Bar({ foo = '' }: { foo: any }) {\n        return foo;\n      }\n    ",
		},
		{
			Code: "\n      function Bar({ foo = '' }: { foo: unknown }) {\n        return foo;\n      }\n    ",
		},
		{
			Code: "\n      function getValue(): undefined;\n      function getValue(box: { value: string }): string;\n      function getValue({ value = '' }: { value?: string } = {}): string | undefined {\n        return value;\n      }\n    ",
		},
		{
			Code: "\n      function getValueObject({ value = '' }: Partial<{ value: string }>) {\n        return value;\n      }\n    ",
		},
		{
			Code: "\n      const { value = 'default' } = someUnknownFunction();\n    ",
		},
		{
			Code: "\n      const [value = 'default'] = someUnknownFunction();\n    ",
		},
		{
			Code: "\n      for (const { value = 'default' } of []) {\n      }\n    ",
		},
		{
			Code: "\n      for (const [value = 'default'] of []) {\n      }\n    ",
		},
		{
			Code: "\n      declare const x: [[number | undefined]];\n      const [[a = 1]] = x;\n    ",
		},
		{
			Code: "\n      function foo(x: string = '') {}\n    ",
		},
		{
			Code: "\n      class C {\n        method(x: string = '') {}\n      }\n    ",
		},
		{
			Code: "\n      const foo = (x: string = '') => {};\n    ",
		},
		{
			Code: "\n      const obj = { ab: { x: 1 } };\n      const {\n        ['a' + 'b']: { x = 1 },\n      } = obj;\n    ",
		},
		{
			Code: "\n      const obj = { ab: 1 };\n      const { ['a' + 'b']: x = 1 } = obj;\n    ",
		},
		{
			Code: "\n      for ([[a = 1]] of []) {\n      }\n    ",
		},
		{
			Code:     "\n        declare const g: Array<string>;\n        const [foo = ''] = g;\n      ",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "\n        declare const g: Record<string, string>;\n        const { foo = '' } = g;\n      ",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code:     "\n        declare const h: { [key: string]: string };\n        const { bar = '' } = h;\n      ",
			TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
		},
		{
			Code: "\n      declare const g: Array<string>;\n      const [foo = ''] = g;\n    ",
		},
		{
			Code: "\n      declare const g: Record<string, string>;\n      const { foo = '' } = g;\n    ",
		},
		{
			Code: "\n      declare const h: { [key: string]: string };\n      const { bar = '' } = h;\n    ",
		},
		{
			Code: "\n      type Merge = boolean | ((incoming: string[]) => void);\n\n      const policy: { merge: Merge } = {\n        merge: (incoming: string[] = []) => {\n          incoming;\n        },\n      };\n    ",
		},
		{
			Code: "\n      const [a, b = ''] = 'somestr'.split('.');\n    ",
		},
		{
			Code: "\n      declare const params: string[];\n      const [c = '123'] = params;\n    ",
		},
		{
			Code: "\n      declare function useCallback<T>(callback: T);\n      useCallback((value: number[] = []) => {});\n    ",
		},
		{
			Code: "\n      declare const tuple: [string];\n      const [a, b = 'default'] = tuple;\n    ",
		},
		{
			Code: "\n      const run = (cb: (...args: unknown[]) => void) => cb();\n      const cb = (p: boolean = true) => null;\n      run(cb);\n      run((p: boolean = true) => null);\n    ",
		},
		{
			Code: "\n      const { a = 'default' } = Math.random() > 0.5 ? { a: 'Hello' } : {};\n    ",
		},
		{
			Code: "\n      const { a = 'default' } =\n        Math.random() > 0.5 ? (Math.random() > 0.5 ? { a: 'Hello' } : {}) : {};\n    ",
		},
		{
			Code: "\n      function findPosts({\n        category,\n        maxResults = 100,\n      }: {\n        category: string;\n        maxResults?: number;\n      }): Promise<string[]> {\n        return Promise.resolve([category, String(maxResults)]);\n      }\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? {} : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? foo : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = foo && { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? { a: 'foo', ...extra } : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? { ...foo } : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const key = Math.random() > 0.5 ? 'a' : 'b';\n      const { a = 'baz' } = cond ? { [key]: 'foo' } : { [key]: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? foo && { a: 'bar' } : { a: 'baz' };\n    ",
		},
		{
			Code: "\n      const obj: unknown = { a: 'bar' };\n      const { a = 'baz' } = cond ? obj : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      const sym = Symbol('a');\n      const { a = 'baz' } = cond ? { [sym]: 'foo' } : { [sym]: 'bar' };\n    ",
		},
		{
			Code: "\n      const { a = 'baz' } = cond ? { [`a${1}`]: 'foo' } : { a: 'bar' };\n    ",
		},
		{
			Code: "\n      class AbstractEntity {\n        public a: string | undefined;\n        public static fromJson<T extends { a: string }>(\n          this: new () => T,\n          { inner = { a: 'test' } }: { inner?: { a: string } },\n        ): T {\n          const entity = new this();\n          entity.a = inner?.a;\n          return entity;\n        }\n      }\n    ",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "\n        function Bar({ foo = '' }: { foo: string }) {\n          return foo;\n        }\n      ",
			Output: []string{
				"\n        function Bar({ foo }: { foo: string }) {\n          return foo;\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    30,
					EndColumn: 32,
				},
			},
		},
		{
			Code: "\n        class C {\n          public method({ foo = '' }: { foo: string }) {\n            return foo;\n          }\n        }\n      ",
			Output: []string{
				"\n        class C {\n          public method({ foo }: { foo: string }) {\n            return foo;\n          }\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      3,
					Column:    33,
					EndColumn: 35,
				},
			},
		},
		{
			Code: "\n        const { 'literal-key': literalKey = 'default' } = { 'literal-key': 'value' };\n      ",
			Output: []string{
				"\n        const { 'literal-key': literalKey } = { 'literal-key': 'value' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    45,
					EndColumn: 54,
				},
			},
		},
		{
			Code: "\n        [1, 2, 3].map((a = 42) => a + 1);\n      ",
			Output: []string{
				"\n        [1, 2, 3].map((a) => a + 1);\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    28,
					EndColumn: 30,
				},
			},
		},
		{
			Code: "\n        function getValue(): undefined;\n        function getValue(box: { value: string }): string;\n        function getValue({ value = '' }: { value: string } = {}): string | undefined {\n          return value;\n        }\n      ",
			Output: []string{
				"\n        function getValue(): undefined;\n        function getValue(box: { value: string }): string;\n        function getValue({ value }: { value: string } = {}): string | undefined {\n          return value;\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      4,
					Column:    37,
					EndColumn: 39,
				},
			},
		},
		{
			Code: "\n        function getValue([value = '']: [string]) {\n          return value;\n        }\n      ",
			Output: []string{
				"\n        function getValue([value]: [string]) {\n          return value;\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    36,
					EndColumn: 38,
				},
			},
		},
		{
			Code: "\n        declare const x: { hello: { world: string } };\n\n        const {\n          hello: { world = '' },\n        } = x;\n      ",
			Output: []string{
				"\n        declare const x: { hello: { world: string } };\n\n        const {\n          hello: { world },\n        } = x;\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      5,
					Column:    28,
					EndColumn: 30,
				},
			},
		},
		{
			Code: "\n        declare const x: { hello: Array<{ world: string }> };\n\n        const {\n          hello: [{ world = '' }],\n        } = x;\n      ",
			Output: []string{
				"\n        declare const x: { hello: Array<{ world: string }> };\n\n        const {\n          hello: [{ world }],\n        } = x;\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      5,
					Column:    29,
					EndColumn: 31,
				},
			},
		},
		{
			Code: "\n        interface B {\n          foo: (b: boolean | string) => void;\n        }\n\n        const h: B = {\n          foo: (b = false) => {},\n        };\n      ",
			Output: []string{
				"\n        interface B {\n          foo: (b: boolean | string) => void;\n        }\n\n        const h: B = {\n          foo: (b) => {},\n        };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      7,
					Column:    21,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "\n        function foo(a = undefined) {}\n      ",
			Output: []string{
				"\n        function foo(a) {}\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessUndefined",
					Line:      2,
					Column:    26,
					EndColumn: 35,
				},
			},
		},
		{
			Code: "\n        const { a = undefined } = {};\n      ",
			Output: []string{
				"\n        const { a } = {};\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessUndefined",
					Line:      2,
					Column:    21,
					EndColumn: 30,
				},
			},
		},
		{
			Code: "\n        const [a = undefined] = [];\n      ",
			Output: []string{
				"\n        const [a] = [];\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessUndefined",
					Line:      2,
					Column:    20,
					EndColumn: 29,
				},
			},
		},
		{
			Code: "\n        function foo({ a = undefined }) {}\n      ",
			Output: []string{
				"\n        function foo({ a }) {}\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessUndefined",
					Line:      2,
					Column:    28,
					EndColumn: 37,
				},
			},
		},
		{
			Code: "\n        function myFunction(p1: string, p2: number | undefined = undefined) {\n          console.log(p1, p2);\n        }\n      ",
			Output: []string{
				"\n        function myFunction(p1: string, p2?: number | undefined) {\n          console.log(p1, p2);\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalSyntax",
					Line:      2,
					Column:    66,
					EndColumn: 75,
				},
			},
		},
		{
			Code: "\n        type SomeType = number | undefined;\n        function f(\n          /* comment */ x /* comment 2 */ : /* comment 3 */ SomeType /* comment 4 */ = /* comment 5 */ undefined,\n        ) {}\n      ",
			Output: []string{
				"\n        type SomeType = number | undefined;\n        function f(\n          /* comment */ x? /* comment 2 */ : /* comment 3 */ SomeType,\n        ) {}\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferOptionalSyntax",
					Line:      4,
					Column:    104,
					EndColumn: 113,
				},
			},
		},
		{
			Code:     "\n        function Bar({ foo = '' }: { foo: string }) {\n          return foo;\n        }\n      ",
			TSConfig: "tsconfig.unstrict.json",
			Output: []string{
				"\n        function Bar({ foo }: { foo: string }) {\n          return foo;\n        }\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
					Line:      0,
					Column:    1,
				},
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    30,
					EndColumn: 32,
				},
			},
		},
		{
			Code:     "\n        function foo(a = undefined) {}\n      ",
			TSConfig: "tsconfig.unstrict.json",
			Output: []string{
				"\n        function foo(a) {}\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
					Line:      0,
					Column:    1,
				},
				{
					MessageId: "uselessUndefined",
					Line:      2,
					Column:    26,
					EndColumn: 35,
				},
			},
		},
		{
			Code: "\n        const { a = 'baz' } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };\n      ",
			Output: []string{
				"\n        const { a } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    21,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "\n        const { a = 'baz' } =\n          Math.random() < 0.5\n            ? { a: 'foo' }\n            : Math.random() > 0.2\n              ? { a: 'bar' }\n              : { a: 'qux' };\n      ",
			Output: []string{
				"\n        const { a } =\n          Math.random() < 0.5\n            ? { a: 'foo' }\n            : Math.random() > 0.2\n              ? { a: 'bar' }\n              : { a: 'qux' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    21,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "\n        const { a = 'baz' } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };\n      ",
			Output: []string{
				"\n        const { a } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    21,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "\n        const { a = 'baz' } = cond ? { a() {} } : { a: 'bar' };\n      ",
			Output: []string{
				"\n        const { a } = cond ? { a() {} } : { a: 'bar' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    21,
					EndColumn: 26,
				},
			},
		},
		{
			Code: "\n        const { a = 'b' } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n      ",
			Output: []string{
				"\n        const { a } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n      ",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "uselessDefaultAssignment",
					Line:      2,
					Column:    21,
					EndColumn: 24,
				},
			},
		},
	})
}
