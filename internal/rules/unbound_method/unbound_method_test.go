package unbound_method

import (
	"slices"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func addContainsMethodsClass(code string) string {
	return `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

let instance = new ContainsMethods();

const arith = {
  double(this: void, x: number): number {
    return x * 2;
  }
};

` + code
}
func addContainsMethodsClassInvalid(code ...string) []rule_tester.InvalidTestCase {
	return utils.Map(code, func(code string) rule_tester.InvalidTestCase {
		return rule_tester.InvalidTestCase{
			Code: addContainsMethodsClass(code),
			// Only: code ==       "const unbound = instance.unbound;",
			Errors: []rule_tester.InvalidTestCaseError{{
				Line:      18,
				MessageId: "unboundWithoutThisAnnotation",
			}},
		}
	})
}

func TestUnboundMethodRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnboundMethodRule, slices.Concat([]rule_tester.ValidTestCase{
		{Code: "Promise.resolve().then(console.log);"},
		{Code: "['1', '2', '3'].map(Number.parseInt);"},
		{Code: "[5.2, 7.1, 3.6].map(Math.floor);"},
		{Code: `
      const foo = Number;
      ['1', '2', '3'].map(foo.parseInt);
    `},
		{Code: `
      const foo = Math;
      [5.2, 7.1, 3.6].map(foo.floor);
    `},
		{Code: "['1', '2', '3'].map(Number['floor']);"},
		{Code: "const x = console.log;"},
		{Code: "const x = Object.defineProperty;"},
		{Code: `
      const foo = Object;
      const x = foo.defineProperty;
    `},
		{Code: "const x = String.fromCharCode;"},
		{Code: `
      const foo = String;
      const x = foo.fromCharCode;
    `},
		{Code: "const x = RegExp.prototype;"},
		{Code: "const x = Symbol.keyFor;"},
		{Code: `
      const foo = Symbol;
      const x = foo.keyFor;
    `},
		{Code: "const x = Array.isArray;"},
		{Code: `
      const foo = Array;
      const x = foo.isArray;
    `},
		{Code: `
      class Foo extends Array {}
      const x = Foo.isArray;
    `},
		{Code: "const x = Proxy.revocable;"},
		{Code: `
      const foo = Proxy;
      const x = foo.revocable;
    `},
		{Code: "const x = Date.parse;"},
		{Code: `
      const foo = Date;
      const x = foo.parse;
    `},
		{Code: "const x = Atomics.load;"},
		{Code: `
      const foo = Atomics;
      const x = foo.load;
    `},
		{Code: "const x = Reflect.deleteProperty;"},
		{Code: "const x = JSON.stringify;"},
		{Code: `
      const foo = JSON;
      const x = foo.stringify;
    `},
		{Code: `
      const o = {
        f: function (this: void) {},
      };
      const f = o.f;
    `},
		// TODO(port): this test passes in tseslint only because there is no DOM lib
		// and  window is an `error` type
		{Skip: true, Code: `
      const { alert } = window;
    `},
		{Code: `
      let b = window.blur;
    `},
		{Code: `
      function foo() {}
      const fooObject = { foo };
      const { foo: bar } = fooObject;
    `},
	},
		utils.Map([]string{
			"instance.bound();",
			"instance.unbound();",

			"ContainsMethods.boundStatic();",
			"ContainsMethods.unboundStatic();",

			"const bound = instance.bound;",
			"const boundStatic = ContainsMethods;",

			"const { bound } = instance;",
			"const { boundStatic } = ContainsMethods;",

			"(instance.bound)();",
			"(instance.unbound)();",

			"(ContainsMethods.boundStatic)();",
			"(ContainsMethods.unboundStatic)();",

			"instance.bound``;",
			"instance.unbound``;",

			"if (instance.bound) { }",
			"if (instance.unbound) { }",

			"if (instance.bound !== undefined) { }",
			"if (instance.unbound !== undefined) { }",

			"if (ContainsMethods.boundStatic) { }",
			"if (ContainsMethods.unboundStatic) { }",

			"if (ContainsMethods.boundStatic !== undefined) { }",
			"if (ContainsMethods.unboundStatic !== undefined) { }",

			"if (ContainsMethods.boundStatic && instance) { }",
			"if (ContainsMethods.unboundStatic && instance) { }",

			"if (instance.bound || instance) { }",
			"if (instance.unbound || instance) { }",

			"ContainsMethods.unboundStatic && 0 || ContainsMethods;",

			"(instance.bound || instance) ? 1 : 0",
			"(instance.unbound || instance) ? 1 : 0",

			"while (instance.bound) { }",
			"while (instance.unbound) { }",

			"while (instance.bound !== undefined) { }",
			"while (instance.unbound !== undefined) { }",

			"while (ContainsMethods.boundStatic) { }",
			"while (ContainsMethods.unboundStatic) { }",

			"while (ContainsMethods.boundStatic !== undefined) { }",
			"while (ContainsMethods.unboundStatic !== undefined) { }",

			"instance.bound as any;",
			"ContainsMethods.boundStatic as any;",

			"instance.bound++;",
			"+instance.bound;",
			"++instance.bound;",
			"instance.bound--;",
			"-instance.bound;",
			"--instance.bound;",
			"instance.bound += 1;",
			"instance.bound -= 1;",
			"instance.bound *= 1;",
			"instance.bound /= 1;",

			"instance.bound || 0;",
			"instance.bound && 0;",

			"instance.bound ? 1 : 0;",
			"instance.unbound ? 1 : 0;",

			"ContainsMethods.boundStatic++;",
			"+ContainsMethods.boundStatic;",
			"++ContainsMethods.boundStatic;",
			"ContainsMethods.boundStatic--;",
			"-ContainsMethods.boundStatic;",
			"--ContainsMethods.boundStatic;",
			"ContainsMethods.boundStatic += 1;",
			"ContainsMethods.boundStatic -= 1;",
			"ContainsMethods.boundStatic *= 1;",
			"ContainsMethods.boundStatic /= 1;",

			"ContainsMethods.boundStatic || 0;",
			"instance.boundStatic && 0;",

			"ContainsMethods.boundStatic ? 1 : 0;",
			"ContainsMethods.unboundStatic ? 1 : 0;",

			"typeof instance.bound === 'function';",
			"typeof instance.unbound === 'function';",

			"typeof ContainsMethods.boundStatic === 'function';",
			"typeof ContainsMethods.unboundStatic === 'function';",

			"instance.unbound = () => {};",
			"instance.unbound = instance.unbound.bind(instance);",
			"if (!!instance.unbound) {}",
			"void instance.unbound",
			"delete instance.unbound",

			"const { double } = arith;",
		}, func(c string) rule_tester.ValidTestCase {
			return rule_tester.ValidTestCase{
				Code: addContainsMethodsClass(c),
				// Only:c  == "if (instance.unbound || instance) { }",
			}
		}),

		[]rule_tester.ValidTestCase{
			{Code: `
interface RecordA {
  readonly type: 'A';
  readonly a: {};
}
interface RecordB {
  readonly type: 'B';
  readonly b: {};
}
type AnyRecord = RecordA | RecordB;

function test(obj: AnyRecord) {
  switch (obj.type) {
  }
}
    `},
			{Code: `
class CommunicationError {
  constructor() {
    const x = CommunicationError.prototype;
  }
}
    `},
			{Code: `
class CommunicationError {}
const x = CommunicationError.prototype;
    `},
			{Code: `
class ContainsMethods {
  bound?: () => void;
  unbound?(): void;

  static boundStatic?: () => void;
  static unboundStatic?(): void;
}

function foo(instance: ContainsMethods | null) {
  instance?.bound();
  instance?.unbound();

  if (instance?.bound) {
  }
  if (instance?.unbound) {
  }

  typeof instance?.bound === 'function';
  typeof instance?.unbound === 'function';
}
    `},
			{Code: `
interface OptionalMethod {
  mightBeDefined?(): void;
}

const x: OptionalMethod = {};
declare const myCondition: boolean;
if (myCondition || x.mightBeDefined) {
  console.log('hello world');
}
    `},
			{Code: `
class A {
  unbound(): void {
    this.unbound = undefined;
    this.unbound = this.unbound.bind(this);
  }
}
    `},
			{Code: "const { parseInt } = Number;"},
			{Code: "const { log } = console;"},
			{Code: `
let parseInt;
({ parseInt } = Number);
    `},
			{Code: `
let log;
({ log } = console);
    `},
			{Code: `
const foo = {
  bar: 'bar',
};
const { bar } = foo;
    `},
			{Code: `
class Foo {
  unbnound() {}
  bar = 4;
}
const { bar } = new Foo();
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
const { bound } = new Foo();
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
function foo({ bound } = new Foo()) {}
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
declare const bar: Foo;
function foo({ bound }: Foo) {}
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo | Bar) {}
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
type foo = ({ bound }: Foo) => void;
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
type foo = ({ unbound }: Foo) => void;
    `},
			{Code: `
class Foo {
  bound = () => 'foo';
}
class Bar {
  bound = () => 'bar';
}
function foo({ bound }: Foo & Bar) {}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
declare const { unbound }: Foo;
    `},
			{Code: "declare const { unbound } = '***';"},
			{Code: `
class Foo {
  unbound = function () {};
}
type foo = (a: (b: (c: ({ unbound }: Foo) => void) => void) => void) => void;
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
class Bar {
  property: ({ unbound }: Foo) => void;
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
function foo<T extends ({ unbound }: Foo) => void>() {}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
abstract class Bar {
  abstract foo({ unbound }: Foo);
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
declare class Bar {
  foo({ unbound }: Foo);
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
declare function foo({ unbound }: Foo);
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo: ({ unbound }: Foo) => void;
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  foo({ unbound }: Foo): void;
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
interface Bar {
  new ({ unbound }: Foo): Foo;
}
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
type foo = new ({ unbound }: Foo) => void;
    `},
			{Code: "const { unbound } = { unbound: () => {} };"},
			{Code: "function foo({ unbound }: { unbound: () => void } = { unbound: () => {} }) {}"},
			{Code: `
class Foo {
  unbound = function () {};
}
declare const foo: Foo

foo.unbound!();
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
declare const foo: Foo

(foo.unbound as any)();
    `},
			{Code: `
class Foo {
  unbound = function () {};
}
declare const foo: Foo

(<any>foo.unbound)();
    `},
			{Code: `
class BaseClass {
  x: number = 42;
  logThis() {}
}
class OtherClass extends BaseClass {
  superLogThis: any;
  constructor() {
    super();
    this.superLogThis = super.logThis;
  }
}
const oc = new OtherClass();
oc.superLogThis();
    `},
		}), slices.Concat([]rule_tester.InvalidTestCase{
		{
			// TODO(port):
			// 2451: Cannot redeclare block-scoped variable 'console'.
			Skip: true,
			Code: `
class Console {
  log(str) {
    process.stdout.write(str);
  }
}

const console = new Console();

Promise.resolve().then(console.log);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unboundWithoutThisAnnotation",
					Line:      10,
				},
			},
		},
		{
			Code: `
import { console } from './class';
const x = console.log;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unboundWithoutThisAnnotation",
					Line:      3,
				},
			},
		},
		{
			Code: addContainsMethodsClass(`
function foo(arg: ContainsMethods | null) {
  const unbound = arg?.unbound;
  arg.unbound += 1;
  arg?.unbound as any;
}
`),
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unboundWithoutThisAnnotation",
					Line:      20,
				},
				{
					MessageId: "unboundWithoutThisAnnotation",
					Line:      21,
				},
				{
					MessageId: "unboundWithoutThisAnnotation",
					Line:      22,
				},
			},
		},
	},
		addContainsMethodsClassInvalid(
			"const unbound = instance.unbound;",
			"const unboundStatic = ContainsMethods.unboundStatic;",

			"const { unbound } = instance;",
			"const { unboundStatic } = ContainsMethods;",

			"<any>instance.unbound;",
			"instance.unbound as any;",

			"<any>ContainsMethods.unboundStatic;",
			"ContainsMethods.unboundStatic as any;",

			"instance.unbound || 0;",
			"ContainsMethods.unboundStatic || 0;",

			"instance.unbound ? instance.unbound : null",
		),
		// here
		[]rule_tester.InvalidTestCase{
			{
				Code: `
class ContainsMethods {
  unbound?(): void;

  static unboundStatic?(): void;
}

new ContainsMethods().unbound;

ContainsMethods.unboundStatic;
      `,
				Options: UnboundMethodOptions{IgnoreStatic: utils.Ref(true)},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      8,
					},
				},
			},
			{
				Code: `
class CommunicationError {
  foo() {}
}
const x = CommunicationError.prototype.foo;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      5,
					},
				},
			},
			{
				Code: "const x = Promise.all;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      1,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound() {}
}
const instance = new Foo();

let x;

x = instance.unbound; // THIS SHOULD ERROR
instance.unbound = x; // THIS SHOULD NOT
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      9,
					},
				},
			},
			{
				Code: `
class Foo extends Number {
  static parseInt = function (string: string, radix?: number): number {};
}
const foo = Foo;
['1', '2', '3'].map(foo.parseInt);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
declare const foo: Number;
const x = foo.toFixed;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
declare const foo: Object;
const x = foo.hasOwnProperty;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
declare const foo: String;
const x = foo.slice;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
declare const foo: Date;
const x = foo.getTime;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo extends Number {}
const x = Foo.parseInt;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo extends String {}
const x = Foo.fromCharCode;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo extends Object {}
const x = Foo.defineProperty;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo extends Date {}
const x = Foo.parse;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
const unbound = new Foo().unbound;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound() {}
}
const { unbound } = new Foo();
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
const { unbound } = new Foo();
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound() {}
}
let unbound;
({ unbound } = new Foo());
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
let unbound;
({ unbound } = new Foo());
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo = new Foo()) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = bar) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: () => {} }) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
declare const bar: Foo;
function foo({ unbound }: Foo = { unbound: function () {} }) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: Foo) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: Foo) => void) {}
bar(({ unbound }) => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
function bar(cb: (arg: { unbound: () => void }) => void) {}
bar(({ unbound } = new Foo()) => {});
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
for (const { unbound } of [new Foo(), new Foo()]) {
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};

  foo({ unbound }: Foo) {}
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = function () {};
}
function foo({ unbound }: Foo | Bar) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      8,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
function foo({ unbound }: { unbound: () => string } | Foo) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
function foo({ unbound }: Foo | Bar) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      8,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
const foo = ({ unbound }: Foo & { foo: () => 'bar' }) => {};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      5,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: (Foo & { foo: () => 'bar' }) | Bar) => {};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      8,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound }: Foo & Bar) => {};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      8,
					},
				},
			},
			{
				Code: `
class Foo {
  unbound = function () {};

  other = function () {};
}
class Bar {
  unbound = () => {};
}
const foo = ({ unbound, ...rest }: Foo & Bar) => {};
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      10,
					},
				},
			},
			{
				Code: "const { unbound } = { unbound: function () {} };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      1,
					},
				},
			},
			{
				Code: `
function foo(
  { unbound }: { unbound: () => void } = { unbound: function () {} },
) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: `
class Foo {
  floor = function () {};
}

const { floor } = Math.random() > 0.5 ? new Foo() : Math;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unbound",
						Line:      6,
					},
				},
			},
			{
				Code: `
class CommunicationError {
  foo() {}
}
const { foo } = CommunicationError.prototype;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      5,
					},
				},
			},
			{
				Code: `
class CommunicationError {
  foo() {}
}
let foo;
({ foo } = CommunicationError.prototype);
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      6,
					},
				},
			},
			{
				Code: `
import { console } from './class';
const { log } = console;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      3,
					},
				},
			},
			{
				Code: "const { all } = Promise;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      1,
					},
				},
			},
			{
				Code: `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    const x = super.logThis;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      8,
						Column:    15,
					},
				},
			},
			{
				Code: `
class BaseClass {
  logThis() {}
}
class OtherClass extends BaseClass {
  constructor() {
    super();
    let x;
    x = super.logThis;
  }
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      9,
						Column:    9,
					},
				},
			},
			{
				Code: `
const values = {
  a() {},
  b: () => {},
};

const { a, b } = values;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      7,
						Column:    9,
						EndColumn: 10,
					},
				},
			},
			{
				Code: `
const values = {
  a() {},
  b: () => {},
};

const { a: c } = values;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      7,
						Column:    9,
						EndColumn: 10,
					},
				},
			},
			{
				Code: `
const values = {
  a() {},
  b: () => {},
};

const { b, a } = values;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      7,
						Column:    12,
						EndColumn: 13,
					},
				},
			},
			{
				Code: `
const objectLiteral = {
  f: function () {},
};
const f = objectLiteral.f;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unboundWithoutThisAnnotation",
						Line:      5,
					},
				},
			},
		}))
}
