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
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  throw new Error('boom');
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo((): ReturnType<typeof foo> => {
  return;
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
type Void = void;
foo((): Void => {
  return;
});
      `,
		},
		{
			Code: `
declare function foo(cb?: () => void): void;
foo();
      `,
		},
		{
			Code: `
declare function foo(...cbs: Array<() => void>): void;
foo(
  () => {},
  () => void null,
  () => undefined,
);
      `,
		},
		{
			Code: `
declare function foo(...cbs: [() => any, () => void, (() => void)?]): void;
foo(
  async () => {},
  () => void null,
  () => undefined,
);
      `,
		},
		{
			Code: `
declare function foo(cb: (() => void) | null): void;
foo(null);
      `,
		},
		{
			Code: `
declare function foo(cb: (() => void) | (() => string)): void;
foo(() => {
  if (maybe) {
    return 'a';
  }
});
      `,
		},
		{
			Code: `
declare const foo: {
  (cb: () => boolean): void;
  (cb: () => void): void;
};
foo(function () {
  with ({}) {
    return false;
  }
});
      `,
		},
		{
			Code: `
declare function cb(): void;
const foo: () => void = cb;
      `,
		},
		{
			Code: `
declare function foo(cb: () => () => void): void;
foo(function () {
  return () => {};
});
      `,
		},
		{
			Code: `
declare function Foo(props: { cb: () => void }): unknown;
const _ = <Foo cb={() => {}} />;
      `,
			Tsx: true,
		},
		{
			Code: `
interface Props {
  cb: (() => void) | (() => Promise<void>);
}
declare function Foo(props: Props): unknown;
const _ = <Foo cb={async () => {}} />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function foo(cb: {}): void;
foo(() => () => []);
      `,
		},
		{
			Code: `
declare function foo(cb: any): void;
foo(() => () => []);
      `,
		},
		{
			Code: `
declare class Foo {
  constructor(cb: unknown): void;
}
new Foo(() => ({}));
      `,
		},
		{
			Code: `
declare function foo(cb: () => Promise<void>): void;
declare function foo(cb: () => void): void;
foo(async () => {});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(cb);
function cb() {
  throw new Error('boom');
}
      `,
		},
		{
			Code: `
declare function foo(...cbs: Array<() => void>): void;
declare const cbs: Array<() => void>;
foo(...cbs);
      `,
		},
		{
			Code: `
let cb;
cb = async () => 10;
      `,
		},
		{
			Code: `
const foo: { (): string; (): void } = () => {
  return 'a';
};
      `,
		},
		{
			Code: `
declare let foo: { cb: (() => void) | number };
foo = {
  cb: 0,
};
      `,
		},
		{
			Code: `
class Foo {
  foo: () => void = () => undefined;
}
      `,
		},
		{
			Code: `
abstract class Foo {
  abstract cb(): void;
}
class Bar extends Foo {
  cb() {
    console.log('a');
  }
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
		{
			Code: `
declare const obj: { foo(cb: () => void): void } | null;
obj?.foo(() => JSON.parse('{}'));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
((cb: () => void) => cb())!(() => 1);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 2},
			},
		},
		{
			Code: `
type AnyFunc = (...args: unknown[]) => unknown;
declare function foo<F extends AnyFunc>(cb: F): void;
foo(async () => ({}));
foo<() => void>(async () => ({}));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 5},
			},
		},
		{
			Code: `
declare function foo<T extends {}>(arg: T, cb: () => T): void;
declare function foo(arg: any, cb: () => void): void;
foo(null, async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  if (maybe) {
    return 1;
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
declare function foo(...cbs: Array<() => void>): void;
foo(
  () => {},
  () => false,
  () => 0,
  () => '',
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
				{MessageId: "nonVoidReturn", Line: 6},
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
const arr = [1, 2];
arr.forEach(async x => {
  console.log(x);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 3},
			},
		},
		{
			Code: `
const foo: () => void = async () => Promise.resolve(true);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `
declare function cb(): unknown;
declare let foo: () => void;
foo = cb;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare function Foo(props: { cb: () => void }): unknown;
const _ = <Foo cb={() => 1} />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
			Tsx: true,
		},
		{
			Code: `
type Cb = () => void;
declare function Foo(props: { cb: Cb; s: string }): unknown;
const _ = <Foo cb={async function () {}} s="test" />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
			Tsx: true,
		},
		{
			Code: `
class Foo {
  cb() {
    console.log('siema');
  }
}
const method = 'cb' as const;
class Bar extends Foo {
  [method]() {
    return 'nara';
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 10},
			},
		},
		{
			Code: `
declare function foo(cb?: { (): void }): void;
foo(() => () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
declare function foo(cb: () => any): void;
foo(async () => {
  return Math.random();
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
		},
		{
			Code: `
interface Cb {
  (arg: string): void;
  (arg: number): void;
}
declare function foo(cb: Cb): void;
foo(cb);
function cb() {
  return true;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 7},
			},
		},
		{
			Code: `
declare function foo(...cbs: [() => void, () => void, (() => void)?]): void;
foo(
  () => {},
  () => Math.random(),
  () => (1).toString(),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
const foo: () => void = () => false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 2},
			},
		},
		{
			Code: `
const { name }: () => void = function foo() {
  return false;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare const foo: Record<string, () => void>;
foo['a' + 'b'] = () => true;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  static foo: () => void = Math.random;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  cb = () => {};
}
class Bar extends Foo {
  cb = Math.random;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 6},
			},
		},
		{
			Code: `
declare let foo: () => () => void;
foo = () => () => 1 + 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function foo(cb: () => () => void): void;
foo(function () {
  return async () => {};
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
		},
	})
}
