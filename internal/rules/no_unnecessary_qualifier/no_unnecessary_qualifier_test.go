package no_unnecessary_qualifier

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUnnecessaryQualifierRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnnecessaryQualifierRule, []rule_tester.ValidTestCase{
		{Code: `
namespace X {
  export type T = number;
}

namespace Y {
  export const x: X.T = 3;
}
    `},
		{Code: `
namespace A {}
namespace A.B {
  export type Z = 1;
}
    `},
		{Code: `
enum A {
  X,
  Y,
}

enum B {
  Z = A.X,
}
    `},
		{Code: `
namespace X {
  export type T = number;
  namespace Y {
    type T = string;
    const x: X.T = 0;
  }
}
    `},
		{Code: `const x: A.B = 3;`},
		{Code: `
namespace X {
  const z = X.y;
}
    `},
		{Code: `
enum Foo {
  One,
}

namespace Foo {
  export function bar() {
    return Foo.One;
  }
}
    `},
		{Code: `
namespace Foo {
  export enum Foo {
    One,
  }
}

namespace Foo {
  export function bar() {
    return Foo.One;
  }
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
namespace A {
  export type B = number;
  const x: A.B = 3;
}
      `,
			Output: []string{`
namespace A {
  export type B = number;
  const x: B = 3;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace A {
  export const x = 3;
  export const y = A.x;
}
      `,
			Output: []string{`
namespace A {
  export const x = 3;
  export const y = x;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace A {
  export type T = number;
  export namespace B {
    const x: A.T = 3;
  }
}
      `,
			Output: []string{`
namespace A {
  export type T = number;
  export namespace B {
    const x: T = 3;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace A {
  export namespace B {
    export type T = number;
    const x: A.B.T = 3;
  }
}
      `,
			Output: []string{`
namespace A {
  export namespace B {
    export type T = number;
    const x: T = 3;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace A {
  export namespace B.C {
    export type D = number;
    const x: A.B.C.D = 3;
  }
}
      `,
			Output: []string{`
namespace A {
  export namespace B.C {
    export type D = number;
    const x: D = 3;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace Outer {
  export type T = number;
  export namespace A.B {}
  const x: Outer.T = 3;
}
      `,
			Output: []string{`
namespace Outer {
  export type T = number;
  export namespace A.B {}
  const x: T = 3;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace A {
  export namespace B {
    export const x = 3;
    const y = A.B.x;
  }
}
      `,
			Output: []string{`
namespace A {
  export namespace B {
    export const x = 3;
    const y = x;
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
enum A {
  B,
  C = A.B,
}
      `,
			Output: []string{`
enum A {
  B,
  C = B,
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
namespace Foo {
  export enum A {
    B,
    C = Foo.A.B,
  }
}
      `,
			Output: []string{`
namespace Foo {
  export enum A {
    B,
    C = B,
  }
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
		{
			Code: `
import * as Foo from './foo';
declare module './foo' {
  const x: Foo.T = 3;
}
      `,
			Output: []string{`
import * as Foo from './foo';
declare module './foo' {
  const x: T = 3;
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryQualifier`}},
		},
	})
}
