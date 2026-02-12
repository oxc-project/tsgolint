package strict_void_return

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestStrictVoidReturnRule(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &StrictVoidReturnRule, []rule_tester.ValidTestCase{
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(function () {
  return;
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => undefined);
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
declare function cb(): void;
foo(cb);
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
declare function boom(): never;
foo(boom);
foo(() => boom());
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => 1 as any);
      `,
			Options: rule_tester.OptionsFromJSON[StrictVoidReturnOptions](`{"allowReturnAny": true}`),
		},
		{
			Code: `
declare const foo: {
  bar(cb1: () => unknown, cb2: () => void): void;
};
foo.bar(
  function () {
    return 1;
  },
  function () {
    return;
  },
);
      `,
		},
		{
			Code: `
declare let foo: { cb?: () => void };
declare function defaultCb(): object;
const { cb = defaultCb } = foo;
      `,
		},
		{
			Code: `
class Foo {
  cb() {
    console.log('a');
  }
}
class Bar extends Foo {
  cb() {
    console.log('b');
  }
}
      `,
		},
		{
			Code: `
interface Foo {
  cb: () => void;
}
class Bar implements Foo {
  cb = () => {};
}
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => null);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function foo(arg: number, cb: () => void): void;
foo(0, () => 0);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function foo(cb: { (): void }): void;
declare function cb(): string;
foo(cb);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare function foo(cb: { (): void }): void;
foo(cb);
async function cb() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(async () => {
  await Promise.resolve();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 3},
			},
		},
		{
			Code: `
const cb: () => void = function* foo() {};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 2},
			},
		},
		{
			Code: `
const cb: () => void = (): Array<number> => [];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 2},
			},
		},
		{
			Code: `
const cb: () => void = (): Array<number> => {
  return [];
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 2},
			},
		},
		{
			Code: `
const foo: () => void = function () {
  if (maybe) {
    return null;
  } else {
    return null;
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 4},
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
declare let foo: { arg?: string; cb?: () => void };
foo.cb = () => {
  return 'siema';
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 4},
			},
		},
		{
			Code: `
declare function cb(): unknown;
let foo: (() => void) | null = null;
foo ??= cb;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare function cb(): unknown;
let foo: (() => void) | boolean = false;
foo ||= cb;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare let foo: { cb: (n: number) => void };
foo = {
  cb(n) {
    return n;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
class Foo {
  cb() {}
}
class Bar extends Foo {
  cb() {
    return Math.random();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
interface Foo {
  cb: () => void;
}
class Bar implements Foo {
  cb = Math.random;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 6},
			},
		},
		{
			Code: `
interface Foo {
  cb(): void;
}
class Bar implements Foo {
  async cb() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 6},
			},
		},
	})
}
