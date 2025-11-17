package no_implied_eval

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoImpliedEvalRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoImpliedEvalRule, []rule_tester.ValidTestCase{
		{Code: "foo.setImmediate(null);"},
		{Code: "foo.setInterval(null);"},
		{Code: "foo.execScript(null);"},
		{Code: "foo.setTimeout(null);"},
		{Code: "foo();"},
		{Code: "(function () {})();"},
		{Code: "setTimeout(() => {}, 0);"},
		{Code: "window.setTimeout(() => {}, 0);"},
		{Code: "window['setTimeout'](() => {}, 0);"},
		{Code: "setInterval(() => {}, 0);"},
		{Code: "window.setInterval(() => {}, 0);"},
		{Code: "window['setInterval'](() => {}, 0);"},
		{Code: "setImmediate(() => {});"},
		{Code: "window.setImmediate(() => {});"},
		{Code: "window['setImmediate'](() => {});"},
		{Code: "execScript(() => {});"},
		{Code: "window.execScript(() => {});"},
		{Code: "window['execScript'](() => {});"},
		{Code: `
const foo = () => {};

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
    `},
		{Code: `
const foo = function () {};

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
    `},
		{Code: `
function foo() {}

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
    `},
		{Code: `
const foo = {
  fn: () => {},
};

setTimeout(foo.fn, 0);
setInterval(foo.fn, 0);
setImmediate(foo.fn);
execScript(foo.fn);
    `},
		{Code: `
const foo = {
  fn: function () {},
};

setTimeout(foo.fn, 0);
setInterval(foo.fn, 0);
setImmediate(foo.fn);
execScript(foo.fn);
    `},
		{Code: `
const foo = {
  fn: function foo() {},
};

setTimeout(foo.fn, 0);
setInterval(foo.fn, 0);
setImmediate(foo.fn);
execScript(foo.fn);
    `},
		{Code: `
const foo = {
  fn() {},
};

setTimeout(foo.fn, 0);
setInterval(foo.fn, 0);
setImmediate(foo.fn);
execScript(foo.fn);
    `},
		{Code: `
const foo = {
  fn: () => {},
};
const fn = 'fn';

setTimeout(foo[fn], 0);
setInterval(foo[fn], 0);
setImmediate(foo[fn]);
execScript(foo[fn]);
    `},
		{Code: `
const foo = {
  fn: () => {},
};

setTimeout(foo['fn'], 0);
setInterval(foo['fn'], 0);
setImmediate(foo['fn']);
execScript(foo['fn']);
    `},
		{Code: `
const foo: () => void = () => {};

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
    `},
		{Code: `
const foo: () => () => void = () => {
  return () => {};
};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
const foo: () => () => void = () => () => {};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
const foo = () => () => {};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
const foo = function foo() {
  return function foo() {};
};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
const foo = function () {
  return function () {
    return '';
  };
};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
const foo: () => () => void = function foo() {
  return function foo() {};
};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
    `},
		{Code: `
function foo() {
  return function foo() {
    return () => {};
  };
}

setTimeout(foo()(), 0);
setInterval(foo()(), 0);
setImmediate(foo()());
execScript(foo()());
    `},
		{Code: `
class Foo {
  static fn = () => {};
}

setTimeout(Foo.fn, 0);
setInterval(Foo.fn, 0);
setImmediate(Foo.fn);
execScript(Foo.fn);
    `},
		{Code: `
class Foo {
  fn() {}
}

const foo = new Foo();

setTimeout(foo.fn, 0);
setInterval(foo.fn, 0);
setImmediate(foo.fn);
execScript(foo.fn);
    `},
		{Code: `
class Foo {
  fn() {}
}
const foo = new Foo();
const fn = foo.fn;

setTimeout(fn.bind(null), 0);
setInterval(fn.bind(null), 0);
setImmediate(fn.bind(null));
execScript(fn.bind(null));
    `},
		{Code: `
const fn = (foo: () => void) => {
  setTimeout(foo, 0);
  setInterval(foo, 0);
  setImmediate(foo);
  execScript(foo);
};
    `},
		{Code: `
import { Function } from './class';
new Function('foo');
    `},
		{Code: `
const foo = (callback: Function) => {
  setTimeout(callback, 0);
};
    `},
		{Code: `
const foo = () => {};
const bar = () => {};

setTimeout(Math.radom() > 0.5 ? foo : bar, 0);
setTimeout(foo || bar, 500);
    `},
		{Code: `
class Foo {
  func1() {}
  func2(): void {
    setTimeout(this.func1.bind(this), 1);
  }
}
    `},
		{Code: `
class Foo {
  private a = {
    b: {
      c: function () {},
    },
  };
  funcw(): void {
    setTimeout(this.a.b.c.bind(this), 1);
  }
}
    `},
		{Code: `
function setTimeout(input: string, value: number) {}

setTimeout('', 0);
    `},
		{Code: `
declare module 'my-timers-promises' {
  export function setTimeout(ms: number): void;
}

import { setTimeout } from 'my-timers-promises';

setTimeout(1000);
    `},
		{Code: `
function setTimeout() {}

{
  setTimeout(100);
}
    `},
		{Code: `
function setTimeout() {}

{
  setTimeout("alert('evil!')");
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
setTimeout('x = 1', 0);
setInterval('x = 1', 0);
setImmediate('x = 1');
execScript('x = 1');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: `
setTimeout(undefined, 0);
setInterval(undefined, 0);
setImmediate(undefined);
execScript(undefined);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: `
setTimeout(1 + '' + (() => {}), 0);
setInterval(1 + '' + (() => {}), 0);
setImmediate(1 + '' + (() => {}));
execScript(1 + '' + (() => {}));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = 'x = 1';

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = function () {
  return 'x + 1';
};

setTimeout(foo(), 0);
setInterval(foo(), 0);
setImmediate(foo());
execScript(foo());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      8,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      9,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = function () {
  return () => 'x + 1';
};

setTimeout(foo()(), 0);
setInterval(foo()(), 0);
setImmediate(foo()());
execScript(foo()());
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      8,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      9,
					Column:    12,
				},
			},
		},
		{
			Code: `
const fn = function () {};

setTimeout(fn + '', 0);
setInterval(fn + '', 0);
setImmediate(fn + '');
execScript(fn + '');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo: string = 'x + 1';

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = new String('x + 1');

setTimeout(foo, 0);
setInterval(foo, 0);
setImmediate(foo);
execScript(foo);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    12,
				},
			},
		},
		{
			Code: `
const foo = 'x + 1';

setTimeout(foo as any, 0);
setInterval(foo as any, 0);
setImmediate(foo as any);
execScript(foo as any);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    12,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    13,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      7,
					Column:    12,
				},
			},
		},
		{
			Code: `
const fn = (foo: string | any) => {
  setTimeout(foo, 0);
  setInterval(foo, 0);
  setImmediate(foo);
  execScript(foo);
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    14,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      4,
					Column:    15,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    16,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    14,
				},
			},
		},
		{
			Code: `
const foo = 'foo';
const bar = () => {};

setTimeout(Math.radom() > 0.5 ? foo : bar, 0);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: `
window.setTimeout(` + "`" + `` + "`" + `, 0);
window['setTimeout'](` + "`" + `` + "`" + `, 0);

window.setInterval(` + "`" + `` + "`" + `, 0);
window['setInterval'](` + "`" + `` + "`" + `, 0);

window.setImmediate(` + "`" + `` + "`" + `);
window['setImmediate'](` + "`" + `` + "`" + `);

window.execScript(` + "`" + `` + "`" + `);
window['execScript'](` + "`" + `` + "`" + `);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    19,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    22,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    20,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    23,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      8,
					Column:    21,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      9,
					Column:    24,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      11,
					Column:    19,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      12,
					Column:    22,
				},
			},
		},
		{
			Code: `
global.setTimeout(` + "`" + `` + "`" + `, 0);
global['setTimeout'](` + "`" + `` + "`" + `, 0);

global.setInterval(` + "`" + `` + "`" + `, 0);
global['setInterval'](` + "`" + `` + "`" + `, 0);

global.setImmediate(` + "`" + `` + "`" + `);
global['setImmediate'](` + "`" + `` + "`" + `);

global.execScript(` + "`" + `` + "`" + `);
global['execScript'](` + "`" + `` + "`" + `);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    19,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    22,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    20,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    23,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      8,
					Column:    21,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      9,
					Column:    24,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      11,
					Column:    19,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      12,
					Column:    22,
				},
			},
		},
		{
			Code: `
globalThis.setTimeout(` + "`" + `` + "`" + `, 0);
globalThis['setTimeout'](` + "`" + `` + "`" + `, 0);

globalThis.setInterval(` + "`" + `` + "`" + `, 0);
globalThis['setInterval'](` + "`" + `` + "`" + `, 0);

globalThis.setImmediate(` + "`" + `` + "`" + `);
globalThis['setImmediate'](` + "`" + `` + "`" + `);

globalThis.execScript(` + "`" + `` + "`" + `);
globalThis['execScript'](` + "`" + `` + "`" + `);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      2,
					Column:    23,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      3,
					Column:    26,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    24,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      6,
					Column:    27,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      8,
					Column:    25,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      9,
					Column:    28,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      11,
					Column:    23,
				},
				{
					MessageId: "noImpliedEvalError",
					Line:      12,
					Column:    26,
				},
			},
		},
		{
			Code: `
const foo: string | undefined = 'hello';
const bar = () => {};

setTimeout(foo || bar, 500);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noImpliedEvalError",
					Line:      5,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = Function();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = new Function('a', 'b', 'return a + b');",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = window.Function();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = new window.Function();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = window['Function']();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "const fn = new window['Function']();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    12,
				},
			},
		},
	})
}
