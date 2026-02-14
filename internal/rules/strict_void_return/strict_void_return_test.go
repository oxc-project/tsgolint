package strict_void_return

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestStrictVoidReturnRule(t *testing.T) {
	t.Parallel()

	validCases := []rule_tester.ValidTestCase{
		{
			Code: `
declare function foo(cb: {}): void;
foo(() => () => []);
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
declare function foo(cb: () => void): void;
foo((): ReturnType<typeof foo> => {
  return;
});
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
declare function foo(cb: () => {}): void;
foo(() => 1 as any);
      `,
			Options: rule_tester.OptionsFromJSON[StrictVoidReturnOptions](`{"allowReturnAny": true}`),
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
declare function boom(): never;
foo(() => boom());
foo(boom);
      `,
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => any): void;
};
new Foo(function () {
  return 1;
});
      `,
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => unknown): void;
};
new Foo(function () {
  return 1;
});
      `,
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
declare const Foo: {
  new (cb: () => string | void): void;
};
new Foo(() => {
  if (maybe) {
    return 'a';
  } else {
    return 'b';
  }
});
      `,
		},
		{
			Code: `
declare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;
foo(() => {
  console.log('a');
});
      `,
		},
		{
			Code: `
declare function foo(cb: (() => void) | (() => string)): void;
foo(() => {
  label: while (maybe) {
    for (let i = 0; i < 10; i++) {
      switch (i) {
        case 0:
          continue;
        case 1:
          return 'a';
      }
    }
  }
});
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
interface Cb {
  (): void;
  (): string;
}
declare const Foo: {
  new (cb: Cb): void;
};
new Foo(() => {
  do {
    try {
      throw 1;
    } catch {
      return 'a';
    }
  } while (maybe);
});
      `,
		},
		{
			Code: `
declare const foo: ((cb: () => boolean) => void) | ((cb: () => void) => void);
foo(() => false);
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
declare const Foo: {
  new (cb: () => void): void;
  (cb: () => unknown): void;
};
Foo(() => false);
      `,
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => any): void;
  (cb: () => void): void;
};
new Foo(() => false);
      `,
		},
		{
			Code: `
declare function foo(cb: () => boolean): void;
declare function foo(cb: () => void): void;
foo(() => false);
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
foo(() => 1 as any);
      `,
			Options: rule_tester.OptionsFromJSON[StrictVoidReturnOptions](`{"allowReturnAny": true}`),
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
const cb = () => {};
foo(cb);
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(function () {});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(cb);
function cb() {}
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
foo(function () {
  return;
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(function () {
  return void 0;
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  return;
});
      `,
		},
		{
			Code: `
declare function foo(cb: () => void): void;
declare function cb(): never;
foo(cb);
      `,
		},
		{
			Code: `
declare class Foo {
  constructor(cb: () => void): any;
}
declare function cb(): void;
new Foo(cb);
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
declare function foo(arg: string, cb: () => void): void;
declare function cb(): undefined;
foo('arg', cb);
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
declare class Foo {
  constructor(cb?: () => void): void;
}
declare function cb(): void;
new Foo(cb);
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
declare function foo(...cbs: Array<() => void>): void;
declare const cbs: Array<() => void>;
foo(...cbs);
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
let cb;
cb = async () => 10;
      `,
		},
		{
			Code: `
const foo: () => void = () => {};
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
const foo: () => void = function () {
  throw new Error('boom');
};
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
const foo: (() => void) | (() => number) = () => {
  return 1;
};
      `,
		},
		{
			Code: `
type Foo = () => void;
const foo: Foo = cb;
function cb() {
  return void null;
}
      `,
		},
		{
			Code: `
interface Foo {
  (): void;
}
const foo: Foo = cb;
function cb() {
  return undefined;
}
      `,
		},
		{
			Code: `
declare function cb(): void;
declare let foo: () => void;
foo = cb;
      `,
		},
		{
			Code: `
declare let foo: () => void;
foo += () => 1;
      `,
		},
		{
			Code: `
declare function defaultCb(): object;
declare let foo: { cb?: () => void };
const { cb = defaultCb } = foo;
      `,
		},
		{
			Code: `
let foo: (() => void) | null = null;
foo &&= null;
      `,
		},
		{
			Code: `
declare function cb(): void;
let foo: (() => void) | boolean = false;
foo ||= cb;
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
declare function Foo(props: { cb: () => void }): unknown;
const _ = <Foo cb="() => {}" />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { cb: () => void }): unknown;
const _ = <Foo cb={} />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { cb: () => void }): unknown;
const _ = <Bar children={<Foo cb={() => {}} />} />;
      `,
			Tsx: true,
		},
		{
			Code: `
type Cb = () => void;
declare function Foo(props: { cb: Cb; s: string }): unknown;
const _ = <Foo cb={function () {}} s="asd" />;
      `,
			Tsx: true,
		},
		{
			Code: `
type Cb = () => void;
declare function Foo(props: { x: number; cb?: Cb }): unknown;
const _ = <Foo x={123} />;
      `,
			Tsx: true,
		},
		{
			Code: `
type Cb = (() => void) | (() => number);
declare function Foo(props: { cb?: Cb }): unknown;
const _ = (
  <Foo
    cb={function (arg) {
      return 123;
    }}
  />
);
      `,
			Tsx: true,
		},
		{
			Code: `
interface Props {
  cb: ((arg: unknown) => void) | boolean;
}
declare function Foo(props: Props): unknown;
const _ = <Foo cb />;
      `,
			Tsx: true,
		},
		{
			Code: `
interface Props {
  cb: (() => void) | (() => Promise<void>);
}
declare function Foo(props: Props): any;
const _ = <Foo cb={async () => {}} />;
      `,
			Tsx: true,
		},
		{
			Code: `
interface Props {
  children: (arg: unknown) => void;
}
declare function Foo(props: Props): unknown;
declare function cb(): void;
const _ = <Foo>{cb}</Foo>;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function foo(cbs: { arg: number; cb: () => void }): void;
foo({ arg: 1, cb: () => undefined });
      `,
		},
		{
			Code: `
declare let foo: { arg?: string; cb: () => void };
foo = {
  cb: () => {
    return something;
  },
};
      `,
			Options: rule_tester.OptionsFromJSON[StrictVoidReturnOptions](`{"allowReturnAny": true}`),
		},
		{
			Code: `
declare let foo: { cb: () => void };
foo = {
  cb() {
    return something;
  },
};
      `,
			Options: rule_tester.OptionsFromJSON[StrictVoidReturnOptions](`{"allowReturnAny": true}`),
		},
		{
			Code: `
declare let foo: { cb: () => void };
foo = {
  cb = () => 1,
};
      `,
		},
		{
			Code: `
declare let foo: { cb: (n: number) => void };
let method = 'cb';
foo = {
  [method](n) {
    return n;
  },
};
      `,
		},
		{
			Code: `
let foo = {
  cb(n) {
    return n;
  },
};
      `,
		},
		{
			Code: `
interface Foo {
  fn(): void;
}
let foo: Foo = {
  cb(n) {
    return n;
  },
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
declare function cb(): void;
const foo: Record<string, () => void> = {
  cb1: cb,
  cb2: cb,
};
      `,
		},
		{
			Code: `
declare function cb(): string;
const foo: Record<string, () => void> = {
  ...cb,
};
      `,
		},
		{
			Code: `
declare function cb(): string;
const foo: Record<string, () => void> = {
  ...cb,
  ...{},
};
      `,
		},
		{
			Code: `
declare function cb(): void;
const foo: Array<(() => void) | false> = [false, cb, () => cb()];
      `,
		},
		{
			Code: `
declare function cb(): void;
const foo: [string, () => void, (() => void)?] = ['asd', cb];
      `,
		},
		{
			Code: `
const foo: { cbs: Array<() => void> | null } = {
  cbs: [
    function () {
      return undefined;
    },
    () => {
      return void 0;
    },
    null,
  ],
};
      `,
		},
		{
			Code: `
const foo: { cb: () => void } = class {
  static cb = () => {};
};
      `,
		},
		{
			Code: `
class Foo {
  foo;
}
      `,
		},
		{
			Code: `
class Bar {
  foo() {}
}
class Foo extends Bar {
  foo();
}
      `,
		},
		{
			Code: `
interface Bar {
  foo(): void;
}
class Foo implements Bar {
  get foo() {
    return new Date();
  }
  set foo() {
    return new Date('wtf');
  }
}
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
class Bar {}
class Foo extends Bar {
  foo = () => 1;
}
      `,
		},
		{
			Code: `
class Foo extends Wtf {
  foo = () => 1;
}
      `,
		},
		{
			Code: `
class Foo extends Wtf {
  [unknown] = () => 1;
}
      `,
		},
		{
			Code: `
class Foo {
  cb = () => {
    console.log('siema');
  };
}
class Bar extends Foo {
  cb = () => {
    console.log('nara');
  };
}
      `,
		},
		{
			Code: `
class Foo {
  cb1 = () => {};
}
class Bar extends Foo {
  cb2() {}
}
class Baz extends Bar {
  cb1 = () => {
    console.log('siema');
  };
  cb2() {
    console.log('nara');
  }
}
      `,
		},
		{
			Code: `
class Foo {
  fn() {
    return 'a';
  }
  cb() {}
}
void class extends Foo {
  cb() {
    if (maybe) {
      console.log('siema');
    } else {
      console.log('nara');
    }
  }
};
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
		{
			Code: `
class Bar implements Foo {
  cb = () => 1;
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
interface Foo {
  cb: () => void;
}
class Bar implements Foo {
  get cb() {
    return () => {};
  }
}
      `,
		},
		{
			Code: `
interface Foo {
  cb(): void;
}
class Bar implements Foo {
  cb() {
    return undefined;
  }
}
      `,
		},
		{
			Code: `
interface Foo1 {
  cb1(): void;
}
interface Foo2 {
  cb2: () => void;
}
class Bar implements Foo1, Foo2 {
  cb1() {}
  cb2() {}
}
      `,
		},
		{
			Code: `
interface Foo1 {
  cb1(): void;
}
interface Foo2 extends Foo1 {
  cb2: () => void;
}
class Bar implements Foo2 {
  cb1() {}
  cb2() {}
}
      `,
		},
		{
			Code: `
declare let foo: () => () => void;
foo = () => () => {};
      `,
		},
		{
			Code: `
declare let foo: { f(): () => void };
foo = {
  f() {
    return () => undefined;
  },
};
function cb() {}
      `,
		},
		{
			Code: `
declare let foo: { f(): () => void };
foo.f = function () {
  return () => {};
};
      `,
		},
		{
			Code: `
declare let foo: () => (() => void) | string;
foo = () => 'asd' + 'zxc';
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
declare function foo(cb: (arg: string) => () => void): void;
declare function foo(cb: (arg: number) => () => boolean): void;
foo((arg: number) => {
  return cb;
});
function cb() {
  return true;
}
      `,
		},
	}

	invalidCases := []rule_tester.InvalidTestCase{
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
declare function foo(cb: () => void): void;
foo(() => (((true))));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  if (maybe) {
    return (((1) + 1));
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
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
declare function foo(cb?: { (): void }): void;
foo(() => () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
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
function foo<T extends {}>(arg: T, cb: () => T);
function foo(arg: null, cb: () => void);
function foo(arg: any, cb: () => any) {}

foo(null, () => Math.random());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
declare function foo<T extends {}>(arg: T, cb: () => T): void;
declare function foo(arg: any, cb: () => void): void;

foo(null, async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 5},
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
declare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;
foo(() => {
  console.log('a');
  return 1;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
function bar<Cb extends () => number>(cb: Cb) {
  foo(cb);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare function foo(cb: { (): void }): void;
const cb = () => dunno;
foo!(cb);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare const foo: {
  (arg: boolean, cb: () => void): void;
};
foo(false, () => Promise.resolve(undefined));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
declare const foo: {
  bar(cb1: () => any, cb2: () => void): void;
};
foo.bar(
  () => Promise.resolve(1),
  () => Promise.resolve(1),
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => void): void;
};
new Foo(async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 5},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  label: while (maybe) {
    for (const i of [1, 2, 3]) {
      if (maybe) return null;
      else return null;
    }
  }
  return void 0;
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {
  do {
    try {
      throw 1;
    } catch (e) {
      return null;
    } finally {
      console.log('finally');
    }
  } while (maybe);
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 8},
			},
		},
		{
			Code: `
declare function foo(cb: () => void): void;
foo(async () => {
  try {
    await Promise.resolve();
  } catch {
    console.error('fail');
  }
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 3},
			},
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => void): void;
  (cb: () => unknown): void;
};
new Foo(() => false);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
declare const Foo: {
  new (cb: () => any): void;
  (cb: () => void): void;
};
Foo(() => false);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
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
declare function foo(
  cb: ((arg: number) => void) | ((arg: string) => void),
): void;
foo(cb);
function cb() {
  return 1 + 1;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 5},
			},
		},
		{
			Code: `
declare function foo(cb: (() => void) | null): void;
declare function cb(): boolean;
foo(cb);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
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
interface Ev {}
interface EvMap {
  DOMContentLoaded: Ev;
}
type EvListOrEvListObj = EvList | EvListObj;
interface EvList {
  (evt: Event): void;
}
interface EvListObj {
  handleEvent(object: Ev): void;
}
interface Win {
  addEventListener<K extends keyof EvMap>(
    type: K,
    listener: (ev: EvMap[K]) => any,
  ): void;
  addEventListener(type: string, listener: EvListOrEvListObj): void;
}
declare const win: Win;
win.addEventListener('DOMContentLoaded', ev => ev);
win.addEventListener('custom', ev => ev);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 21},
				{MessageId: "nonVoidReturn", Line: 22},
			},
		},
		{
			Code: `
declare function foo(x: null, cb: () => void): void;
declare function foo(x: unknown, cb: () => any): void;
foo({}, async () => {});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
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
[1, 2].forEach(async x => console.log(x));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
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
const foo: () => void = async () => Promise.resolve(true);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `const cb: () => void = (): Array<number> => [];`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 1},
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
			Code: `const cb: () => void = function* foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 1},
			},
		},
		{
			Code: `const cb: () => void = (): Promise<number> => Promise.resolve(1);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 1},
			},
		},
		{
			Code: `
const cb: () => void = async (): Promise<number> => {
  try {
    return Promise.resolve(1);
  } catch {}
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `const cb: () => void = async (): Promise<number> => Promise.resolve(1);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 1},
			},
		},
		{
			Code: `
const foo: () => void = async () => {
  try {
    return 1;
  } catch {}
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `
const foo: () => void = async (): Promise<void> => {
  try {
    await Promise.resolve();
  } finally {
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `
const foo: () => void = async () => {
  try {
    await Promise.resolve();
  } catch (err) {
    console.error(err);
  }
  console.log('ok');
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `const foo: () => void = (): number => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 1},
			},
		},
		{
			Code: `
declare function cb(): boolean;
const foo: () => void = cb;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
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
const foo: () => void = function () {
  if (maybe) {
    console.log('elo');
    return { [1]: Math.random() };
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
const foo: { (arg: number): void; (arg: string): void } = arg => {
  console.log('foo');
  switch (typeof arg) {
    case 'number':
      return 0;
    case 'string':
      return '';
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
				{MessageId: "nonVoidReturn", Line: 8},
			},
		},
		{
			Code: `
const foo: ((arg: number) => void) | ((arg: string) => void) = async () => {
  return 1;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 2},
			},
		},
		{
			Code: `
type Foo = () => void;
const foo: Foo = cb;
function cb() {
  return [1, 2, 3];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
			},
		},
		{
			Code: `
interface Foo {
  (): void;
}
const foo: Foo = cb;
function cb() {
  return { a: 1 };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 5},
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
declare let foo: { arg?: string; cb?: () => void };
foo.cb = () => {
  return 'siema';
  console.log('siema');
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
declare function cb(): unknown;
let foo: (() => void) | boolean = false;
foo &&= cb;
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
declare function Foo(props: { cb: () => void }): unknown;
declare function getNull(): null;
const _ = (
  <Foo
    cb={() => {
      if (maybe) return Math.random();
      else return getNull();
    }}
  />
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 7},
				{MessageId: "nonVoidReturn", Line: 8},
			},
			Tsx: true,
		},
		{
			Code: `
type Cb = () => void;
declare function Foo(props: { cb: Cb; s: string }): unknown;
const _ = <Foo cb={async function () {}} s="!@#jp2gmd" />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
			Tsx: true,
		},
		{
			Code: `
type Cb = () => void;
declare function Foo(props: { n: number; cb?: Cb }): unknown;
const _ = <Foo n={2137} cb={function* () {}} />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
			Tsx: true,
		},
		{
			Code: `
type Cb = ((arg: string) => void) | ((arg: number) => void);
declare function Foo(props: { cb?: Cb }): unknown;
const _ = (
  <Foo
    cb={async function* (arg) {
      await arg;
      yield arg;
    }}
  />
);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 6},
			},
			Tsx: true,
		},
		{
			Code: `
interface Props {
  cb: ((arg: unknown) => void) | boolean;
}
declare function Foo(props: Props): unknown;
const _ = <Foo cb={x => x} />;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
			},
			Tsx: true,
		},
		{
			Code: `
type EventHandler<E> = { bivarianceHack(event: E): void }['bivarianceHack'];
interface ButtonProps {
  onClick?: EventHandler<unknown> | undefined;
}
declare function Button(props: ButtonProps): unknown;
function App() {
  return <Button onClick={x => x} />;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 8},
			},
			Tsx: true,
		},
		{
			Code: `
declare function foo(cbs: { arg: number; cb: () => void }): void;
foo({ arg: 1, cb: () => 1 });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare let foo: { arg?: string; cb: () => void };
foo = {
  cb: () => {
    let x = 'siema';
    return x;
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
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
declare let foo: { 1234: (n: number) => void };
foo = {
  1234(n) {
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
declare let foo: { '1e+21': () => void };
foo = {
  1_000_000_000_000_000_000_000: () => 1,
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 4},
			},
		},
		{
			Code: `
declare let foo: { cb: (() => void) | number };
foo = {
  cb: async () => {
    if (maybe) {
      return 'asd';
    }
  },
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 4},
			},
		},
		{
			Code: `
declare function cb(): number;
const foo: Record<string, () => void> = {
  cb1: cb,
  cb2: cb,
  ...cb,
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
				{MessageId: "nonVoidFunc", Line: 5},
			},
		},
		{
			Code: `
declare function cb(): number;
const foo: Array<(() => void) | false> = [false, cb, () => cb()];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare function cb(): number;
const foo: [string, () => void, (() => void)?] = ['asd', cb];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 3},
			},
		},
		{
			Code: `
const foo: { cbs: Array<() => void> | null } = {
  cbs: [
    function* () {
      yield 1;
    },
    async () => {
      await 1;
    },
    null,
  ],
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
				{MessageId: "asyncFunc", Line: 7},
			},
		},
		{
			Code: `
const foo: { cb: () => void } = class {
  static cb = () => ({});
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  foo: () => void = () => [];
}
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
const foo = () =>
  class {
    cb = () => {};
  };
class Bar extends foo() {
  cb = Math.random;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 7},
			},
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
			// TSGolint rule doesn't check getter override return types for void methods
			Skip: true,
			Code: `
class Bar {
  foo() {}
}
class Foo extends Bar {
  get foo() {
    return () => 1;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
class Foo {
  cb() {}
}
void class extends Foo {
  cb() {
    return Math.random();
  }
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 7},
			},
		},
		{
			Code: `
class Foo {
  cb1 = () => {};
}
class Bar extends Foo {
  cb2() {}
}
class Baz extends Bar {
  cb1 = () => Math.random();
  cb2() {
    return Math.random();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 9},
				{MessageId: "nonVoidReturn", Line: 11},
			},
		},
		{
			Code: `
declare function f(): Promise<void>;
interface Foo {
  cb: () => void;
}
class Bar {
  cb = () => {};
}
class Baz extends Bar implements Foo {
  cb: () => void = f;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 10},
			},
		},
		{
			Code: `
class Foo {
  fn() {
    return 'a';
  }
  cb() {}
}
class Bar extends Foo {
  cb() {
    if (maybe) {
      return Promise.resolve('siema');
    } else {
      return Promise.resolve('nara');
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 11},
				{MessageId: "nonVoidReturn", Line: 13},
			},
		},
		{
			Code: `
abstract class Foo {
  abstract cb(): void;
}
class Bar extends Foo {
  async cb() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 6},
			},
		},
		{
			Code: `
class Foo {
  fn() {
    return 'a';
  }
  cb() {}
}
class Bar extends Foo {
  *cb() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 9},
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
const o = { cb() {} };
type O = typeof o;
class Bar implements O {
  cb = Math.random;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 5},
			},
		},
		{
			Code: `
class Foo {
  cb() {}
}
class Bar extends Foo {
  async*cb() {}
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
  async cb(): Promise<string> {
    return Promise.resolve('siema');
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 6},
			},
		},
		{
			Code: `
interface Foo {
  cb(): void;
}
class Bar implements Foo {
  async cb() {
    try {
      return { a: ['asdf', 1234] };
    } catch {
      console.error('error');
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 6},
			},
		},
		{
			Code: `
interface Foo {
  cb(): void;
}
class Bar implements Foo {
  cb() {
    if (maybe) {
      return Promise.resolve(1);
    } else {
      return;
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 8},
			},
		},
		{
			Code: `
interface Foo1 {
  cb1(): void;
}
interface Foo2 {
  cb2: () => void;
}
class Bar implements Foo1, Foo2 {
  async cb1() {}
  async *cb2() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 9},
				{MessageId: "nonVoidFunc", Line: 10},
			},
		},
		{
			Code: `
interface Foo1 {
  cb1(): void;
}
interface Foo2 {
  cb2: () => void;
}
class Baz {
  cb3() {}
}
class Bar extends Baz implements Foo1, Foo2 {
  async cb1() {}
  async *cb2() {}
  cb3() {
    return Math.random();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 12},
				{MessageId: "nonVoidFunc", Line: 13},
				{MessageId: "nonVoidReturn", Line: 15},
			},
		},
		{
			Code: `
class A extends class {
  cb() {}
} {
  cb() {
    return Math.random();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
class A extends class B {
  cb() {}
} {
  cb() {
    return Math.random();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
			},
		},
		{
			Code: `
interface Foo1 {
  cb1(): void;
}
interface Foo2 extends Foo1 {
  cb2: () => void;
}
class Bar implements Foo2 {
  async cb1() {}
  async *cb2() {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "asyncFunc", Line: 9},
				{MessageId: "nonVoidFunc", Line: 10},
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
declare let foo: () => () => void;
foo = () => () => Math.random();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 3},
			},
		},
		{
			Code: `
declare let foo: () => () => void;
declare const cb: () => null | false;
foo = () => cb;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 4},
			},
		},
		{
			Code: `
declare let foo: { f(): () => void };
foo = {
  f() {
    return () => cb;
  },
};
function cb() {}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
declare let foo: { f(): () => void };
foo.f = function () {
  return () => {
    return null;
  };
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 5},
			},
		},
		{
			Code: `
declare let foo: () => (() => void) | string;
foo = () => () => {
  return 'asd' + 'zxc';
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 4},
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
		{
			Code: `
declare function foo(cb: () => () => void): void;
foo(() => () => {
  if (n == 1) {
    console.log('asd')
    return [1].map(x => x)
  }
  if (n == 2) {
    console.log('asd')
    return -Math.random()
  }
  if (n == 3) {
    console.log('asd')
    return "x".toUpperCase()
  }
  return <i>{Math.random()}</i>
});
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidReturn", Line: 6},
				{MessageId: "nonVoidReturn", Line: 10},
				{MessageId: "nonVoidReturn", Line: 14},
				{MessageId: "nonVoidReturn", Line: 16},
			},
			Tsx: true,
		},
		{
			Code: `
declare function foo(cb: (arg: string) => () => void): void;
declare function foo(cb: (arg: number) => () => boolean): void;
foo((arg: string) => {
  return cb;
});
async function* cb() {
  yield true;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonVoidFunc", Line: 5},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &StrictVoidReturnRule, validCases, invalidCases)
}
