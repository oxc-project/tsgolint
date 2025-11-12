package prefer_return_this_type

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferReturnThisTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferReturnThisTypeRule, []rule_tester.ValidTestCase{
		{Code: `
class Foo {
  f1() {}
  f2(): Foo {
    return new Foo();
  }
  f3() {
    return this;
  }
  f4(): this {
    return this;
  }
  f5(): any {
    return this;
  }
  f6(): unknown {
    return this;
  }
  f7(foo: Foo): Foo {
    return Math.random() > 0.5 ? foo : this;
  }
  f10(this: Foo, that: Foo): Foo;
  f11(): Foo {
    return;
  }
  f13(this: Foo): Foo {
    return this;
  }
  f14(): { f14: Function } {
    return this;
  }
  f15(): Foo | this {
    return Math.random() > 0.5 ? new Foo() : this;
  }
}
    `},
		{Code: `
class Foo {
  f1 = () => {};
  f2 = (): Foo => {
    return new Foo();
  };
  f3 = () => this;
  f4 = (): this => {
    return this;
  };
  f5 = (): Foo => new Foo();
  f6 = '';
}
    `},
		{Code: `
const Foo = class {
  bar() {
    return this;
  }
};
    `},
		{Code: `
class Base {}
class Derived extends Base {
  f(): Base {
    return this;
  }
}
    `},
		{Code: `
class Foo {
  accessor f = () => {
    return this;
  };
}
    `},
		{Code: `
class Foo {
  accessor f = (): this => {
    return this;
  };
}
    `},
		{Code: `
class Foo {
  f?: string;
}
    `},
		{Code: `
const Foo = class {
	bar(): this {
    return this;
  }
};
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
class Foo {
  f(): Foo {
    return this;
  }
}
      `,
			Output: []string{`
class Foo {
  f(): this {
    return this;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    8,
				},
			},
		},
		{
			Code: `
class Foo {
  f = function (): Foo {
    return this;
  };
}
      `,
			Output: []string{`
class Foo {
  f = function (): this {
    return this;
  };
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
class Foo {
  f(): Foo {
    const self = this;
    return self;
  }
}
      `,
			Output: []string{`
class Foo {
  f(): this {
    const self = this;
    return self;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    8,
				},
			},
		},
		{
			Code: `
class Foo {
  f = (): Foo => {
    return this;
  };
}
      `,
			Output: []string{`
class Foo {
  f = (): this => {
    return this;
  };
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
class Foo {
  f = (): Foo => {
    const self = this;
    return self;
  };
}
      `,
			Output: []string{`
class Foo {
  f = (): this => {
    const self = this;
    return self;
  };
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
class Foo {
  f = (): Foo => this;
}
      `,
			Output: []string{`
class Foo {
  f = (): this => this;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code: `
class Foo {
  accessor f = (): Foo => {
    return this;
  };
}
      `,
			Output: []string{`
class Foo {
  accessor f = (): this => {
    return this;
  };
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
class Foo {
  accessor f = (): Foo => this;
}
      `,
			Output: []string{`
class Foo {
  accessor f = (): this => this;
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
class Foo {
  f1(): Foo | undefined {
    return this;
  }
  f2(): this | undefined {
    return this;
  }
}
      `,
			Output: []string{`
class Foo {
  f1(): this | undefined {
    return this;
  }
  f2(): this | undefined {
    return this;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    9,
				},
			},
		},
		{
			Code: `
class Foo {
  bar(): Foo | undefined {
    if (Math.random() > 0.5) {
      return this;
    }
  }
}
      `,
			Output: []string{`
class Foo {
  bar(): this | undefined {
    if (Math.random() > 0.5) {
      return this;
    }
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    10,
				},
			},
		},
		{
			Code: `
class Foo {
  bar(num: 1 | 2): Foo {
    switch (num) {
      case 1:
        return this;
      case 2:
        return this;
    }
  }
}
      `,
			Output: []string{`
class Foo {
  bar(num: 1 | 2): this {
    switch (num) {
      case 1:
        return this;
      case 2:
        return this;
    }
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    20,
				},
			},
		},
		{
			Code: `
class Animal<T> {
  eat(): Animal<T> {
    console.log("I'm moving!");
    return this;
  }
}
      `,
			Output: []string{`
class Animal<T> {
  eat(): this {
    console.log("I'm moving!");
    return this;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    10,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
const Foo = class FooClass {
  f(): FooClass {
    return this;
  }
}
      `,
			Output: []string{`
const Foo = class FooClass {
  f(): this {
    return this;
  }
}
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useThisType",
					Line:      3,
					Column:    8,
				},
			},
		},
	})
}
