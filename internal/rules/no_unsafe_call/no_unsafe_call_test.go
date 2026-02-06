package no_unsafe_call

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnsafeCallRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noImplicitThis.json", t, &NoUnsafeCallRule, []rule_tester.ValidTestCase{
		{Code: `
function foo(x: () => void) {
  x();
}
    `},
		{Code: `
function foo(x?: { a: () => void }) {
  x?.a();
}
    `},
		{Code: `
function foo(x: { a?: () => void }) {
  x.a?.();
}
    `},
		{Code: "new Map();"},
		{Code: "String.raw`foo`;"},
		{Code: "const x = import('./foo');"},
		{Code: `
      let foo: any = 23;
      String(foo); // ERROR: Unsafe call of an any typed value
    `},
		{Code: `
      function foo<T extends any>(x: T) {
        x();
      }
    `},
		{Code: `
      // create a scope since it's illegal to declare a duplicate identifier
      // 'Function' in the global script scope.
      {
        type Function = () => void;
        const notGlobalFunctionType: Function = (() => {}) as Function;
        notGlobalFunctionType();
      }
    `},
		{Code: `
interface SurprisinglySafe extends Function {
  (): string;
}
declare const safe: SurprisinglySafe;
safe();
    `},
		{Code: `
interface CallGoodConstructBad extends Function {
  (): void;
}
declare const safe: CallGoodConstructBad;
safe();
    `},
		{Code: `
interface ConstructSignatureMakesSafe extends Function {
  new (): ConstructSignatureMakesSafe;
}
declare const safe: ConstructSignatureMakesSafe;
new safe();
    `},
		{Code: `
interface SafeWithNonVoidCallSignature extends Function {
  (): void;
  (x: string): string;
}
declare const safe: SafeWithNonVoidCallSignature;
safe();
    `},
		{Code: `
      new Function('lol');
    `},
		{Code: `
      Function('lol');
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
function foo(x: any) {
  x();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 4,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x?.();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 4,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x.a.b.c.d.e.f.g();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x.a.b.c.d.e.f.g?.();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 18,
				},
			},
		},
		{
			Code: `
function foo(x: { a: any }) {
  x.a();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x: { a: any }) {
  x?.a();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 7,
				},
			},
		},
		{
			Code: `
function foo(x: { a: any }) {
  x.a?.();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    3,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  new x();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeNew",
				},
			},
		},
		{
			Code: `
function foo(x: { a: any }) {
  new x.a();
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeNew",
				},
			},
		},
		{
			Code: `
function foo(x: any) {
  x` + "`" + `foo` + "`" + `;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTemplateTag",
				},
			},
		},
		{
			Code: `
function foo(x: { tag: any }) {
  x.tag` + "`" + `foo` + "`" + `;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTemplateTag",
				},
			},
		},
		{
			Code: `
const methods = {
  methodA() {
    return this.methodB()
  },
  methodB() {
    return true
  },
  methodC() {
    return this()
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCallThis",
					Line:      4,
					Column:    12,
					EndColumn: 24,
				},
				{
					MessageId: "unsafeCallThis",
					Line:      10,
					Column:    12,
					EndColumn: 16,
				},
			},
		},
		{
			Code: `
let value: NotKnown;
value();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
					Column:    1,
					EndColumn: 6,
				},
			},
		},
		{
			Code: `
const t: Function = () => {};
t();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      3,
				},
			},
		},
		{
			Code: `
const f: Function = () => {};
f` + "`" + `oo` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTemplateTag",
					Line:      3,
				},
			},
		},
		{
			Code: `
declare const maybeFunction: unknown;
if (typeof maybeFunction === 'function') {
  maybeFunction('call', 'with', 'any', 'args');
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      4,
				},
			},
		},
		{
			Code: `
interface Unsafe extends Function {}
declare const unsafe: Unsafe;
unsafe();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      4,
				},
			},
		},
		{
			Code: `
interface Unsafe extends Function {}
declare const unsafe: Unsafe;
unsafe` + "`" + `bad` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeTemplateTag",
					Line:      4,
				},
			},
		},
		{
			Code: `
interface Unsafe extends Function {}
declare const unsafe: Unsafe;
new unsafe();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeNew",
					Line:      4,
				},
			},
		},
		{
			Code: `
interface UnsafeToConstruct extends Function {
  (): void;
}
declare const unsafe: UnsafeToConstruct;
new unsafe();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeNew",
					Line:      6,
				},
			},
		},
		{
			Code: `
interface StillUnsafe extends Function {
  property: string;
}
declare const unsafe: StillUnsafe;
unsafe();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeCall",
					Line:      6,
				},
			},
		},
	})
}
