package prefer_readonly

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferReadonlyRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferReadonlyRule, []rule_tester.ValidTestCase{
		{Code: `function ignore() {}`},
		{Code: `const ignore = function () {};`},
		{Code: `const ignore = () => {};`},
		{Code: `
      const container = { member: true };
      container.member;
    `},
		{Code: `
      const container = { member: 1 };
      +container.member;
    `},
		{Code: `
      const container = { member: 1 };
      ++container.member;
    `},
		{Code: `
      const container = { member: 1 };
      container.member++;
    `},
		{Code: `
      const container = { member: 1 };
      -container.member;
    `},
		{Code: `
      const container = { member: 1 };
      --container.member;
    `},
		{Code: `
      const container = { member: 1 };
      container.member--;
    `},
		{Code: `class TestEmpty {}`},
		{Code: `
      class TestReadonlyStatic {
        private static readonly correctlyReadonlyStatic = 7;
      }
    `},
		{Code: `
      class TestReadonlyStatic {
        static readonly #correctlyReadonlyStatic = 7;
      }
    `},
		{Code: `
      class TestModifiableStatic {
        private static correctlyModifiableStatic = 7;

        public constructor() {
          TestModifiableStatic.correctlyModifiableStatic += 1;
        }
      }
    `},
		{Code: `
      class TestModifiableStatic {
        static #correctlyModifiableStatic = 7;

        public constructor() {
          TestModifiableStatic.#correctlyModifiableStatic += 1;
        }
      }
    `},
		{Code: `
      class TestModifiableByParameterProperty {
        private static readonly correctlyModifiableByParameterProperty = 7;

        public constructor(
          public correctlyModifiablePublicParameter: number = (() => {
            return (TestModifiableStatic.correctlyModifiableByParameterProperty += 1);
          })(),
        ) {}
      }
    `},
		{Code: `
      class TestModifiableByParameterProperty {
        static readonly #correctlyModifiableByParameterProperty = 7;

        public constructor(
          public correctlyModifiablePublicParameter: number = (() => {
            return (TestModifiableStatic.#correctlyModifiableByParameterProperty += 1);
          })(),
        ) {}
      }
    `},
		{Code: `
      class TestReadonlyInline {
        private readonly correctlyReadonlyInline = 7;
      }
    `},
		{Code: `
      class TestReadonlyInline {
        readonly #correctlyReadonlyInline = 7;
      }
    `},
		{Code: `
      class TestReadonlyDelayed {
        private readonly correctlyReadonlyDelayed = 7;

        public constructor() {
          this.correctlyReadonlyDelayed += 1;
        }
      }
    `},
		{Code: `
      class TestReadonlyDelayed {
        readonly #correctlyReadonlyDelayed = 7;

        public constructor() {
          this.#correctlyReadonlyDelayed += 1;
        }
      }
    `},
		{Code: `
      class TestModifiableInline {
        private correctlyModifiableInline = 7;

        public mutate() {
          this.correctlyModifiableInline += 1;

          return class {
            private correctlyModifiableInline = 7;

            mutate() {
              this.correctlyModifiableInline += 1;
            }
          };
        }
      }
    `},
		{Code: `
      class TestModifiableInline {
        #correctlyModifiableInline = 7;

        public mutate() {
          this.#correctlyModifiableInline += 1;

          return class {
            #correctlyModifiableInline = 7;

            mutate() {
              this.#correctlyModifiableInline += 1;
            }
          };
        }
      }
    `},
		{Code: `
      class TestModifiableDelayed {
        private correctlyModifiableDelayed = 7;

        public mutate() {
          this.correctlyModifiableDelayed += 1;
        }
      }
    `},
		{Code: `
      class TestModifiableDelayed {
        #correctlyModifiableDelayed = 7;

        public mutate() {
          this.#correctlyModifiableDelayed += 1;
        }
      }
    `},
		{Code: `
      class TestModifiableDeleted {
        private correctlyModifiableDeleted = 7;

        public mutate() {
          delete this.correctlyModifiableDeleted;
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructor {
        private correctlyModifiableWithinConstructor = 7;

        public constructor() {
          (() => {
            this.correctlyModifiableWithinConstructor += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructor {
        #correctlyModifiableWithinConstructor = 7;

        public constructor() {
          (() => {
            this.#correctlyModifiableWithinConstructor += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorArrowFunction {
        private correctlyModifiableWithinConstructorArrowFunction = 7;

        public constructor() {
          (() => {
            this.correctlyModifiableWithinConstructorArrowFunction += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorArrowFunction {
        #correctlyModifiableWithinConstructorArrowFunction = 7;

        public constructor() {
          (() => {
            this.#correctlyModifiableWithinConstructorArrowFunction += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInFunctionExpression {
        private correctlyModifiableWithinConstructorInFunctionExpression = 7;

        public constructor() {
          const self = this;

          (() => {
            self.correctlyModifiableWithinConstructorInFunctionExpression += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInFunctionExpression {
        #correctlyModifiableWithinConstructorInFunctionExpression = 7;

        public constructor() {
          const self = this;

          (() => {
            self.#correctlyModifiableWithinConstructorInFunctionExpression += 1;
          })();
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInGetAccessor {
        private correctlyModifiableWithinConstructorInGetAccessor = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            get accessor() {
              return (self.correctlyModifiableWithinConstructorInGetAccessor += 1);
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInGetAccessor {
        #correctlyModifiableWithinConstructorInGetAccessor = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            get accessor() {
              return (self.#correctlyModifiableWithinConstructorInGetAccessor += 1);
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInMethodDeclaration {
        private correctlyModifiableWithinConstructorInMethodDeclaration = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            methodDeclaration() {
              self.correctlyModifiableWithinConstructorInMethodDeclaration = 7;
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInMethodDeclaration {
        #correctlyModifiableWithinConstructorInMethodDeclaration = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            methodDeclaration() {
              self.#correctlyModifiableWithinConstructorInMethodDeclaration = 7;
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInSetAccessor {
        private correctlyModifiableWithinConstructorInSetAccessor = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            set accessor(value: number) {
              self.correctlyModifiableWithinConstructorInSetAccessor += value;
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiableWithinConstructorInSetAccessor {
        #correctlyModifiableWithinConstructorInSetAccessor = 7;

        public constructor() {
          const self = this;

          const confusingObject = {
            set accessor(value: number) {
              self.#correctlyModifiableWithinConstructorInSetAccessor += value;
            },
          };
        }
      }
    `},
		{Code: `
      class TestModifiablePostDecremented {
        private correctlyModifiablePostDecremented = 7;

        public mutate() {
          this.correctlyModifiablePostDecremented -= 1;
        }
      }
    `},
		{Code: `
      class TestModifiablePostDecremented {
        #correctlyModifiablePostDecremented = 7;

        public mutate() {
          this.#correctlyModifiablePostDecremented -= 1;
        }
      }
    `},
		{Code: `
      class TestyModifiablePostIncremented {
        private correctlyModifiablePostIncremented = 7;

        public mutate() {
          this.correctlyModifiablePostIncremented += 1;
        }
      }
    `},
		{Code: `
      class TestyModifiablePostIncremented {
        #correctlyModifiablePostIncremented = 7;

        public mutate() {
          this.#correctlyModifiablePostIncremented += 1;
        }
      }
    `},
		{Code: `
      class TestModifiablePreDecremented {
        private correctlyModifiablePreDecremented = 7;

        public mutate() {
          --this.correctlyModifiablePreDecremented;
        }
      }
    `},
		{Code: `
      class TestModifiablePreDecremented {
        #correctlyModifiablePreDecremented = 7;

        public mutate() {
          --this.#correctlyModifiablePreDecremented;
        }
      }
    `},
		{Code: `
      class TestModifiablePreIncremented {
        private correctlyModifiablePreIncremented = 7;

        public mutate() {
          ++this.correctlyModifiablePreIncremented;
        }
      }
    `},
		{Code: `
      class TestModifiablePreIncremented {
        #correctlyModifiablePreIncremented = 7;

        public mutate() {
          ++this.#correctlyModifiablePreIncremented;
        }
      }
    `},
		{Code: `
      class TestProtectedModifiable {
        protected protectedModifiable = 7;
      }
    `},
		{Code: `
      class TestPublicModifiable {
        public publicModifiable = 7;
      }
    `},
		{Code: `
      class TestReadonlyParameter {
        public constructor(private readonly correctlyReadonlyParameter = 7) {}
      }
    `},
		{Code: `
      class TestCorrectlyModifiableParameter {
        public constructor(private correctlyModifiableParameter = 7) {}

        public mutate() {
          this.correctlyModifiableParameter += 1;
        }
      }
    `},
		{Code: `
        class TestCorrectlyNonInlineLambdas {
          private correctlyNonInlineLambda = 7;
        }
      `, Options: rule_tester.OptionsFromJSON[PreferReadonlyOptions](`{"onlyInlineLambdas":true}`)},
		{Code: `
        class TestCorrectlyNonInlineLambdas {
          #correctlyNonInlineLambda = 7;
        }
      `, Options: rule_tester.OptionsFromJSON[PreferReadonlyOptions](`{"onlyInlineLambdas":true}`)},
		{Code: `
        class TestCorrectlyNoInitializer {
          private correctlyNoInitializer: number;
        }
      `, Options: rule_tester.OptionsFromJSON[PreferReadonlyOptions](`{"onlyInlineLambdas":true}`)},
		{Code: `
        class TestCorrectlyNoInitializer {
          #correctlyNoInitializer: number;
        }
      `, Options: rule_tester.OptionsFromJSON[PreferReadonlyOptions](`{"onlyInlineLambdas":true}`)},
		{Code: `
      class TestComputedParameter {
        public mutate() {
          this['computed'] = 1;
        }
      }
    `},
		{Code: `
      class TestComputedParameter {
        private ['computed-ignored-by-rule'] = 1;
      }
    `},
		{Code: `
class Foo {
  private value: number = 0;

  bar(newValue: { value: number }) {
    ({ value: this.value } = newValue);
    return this.value;
  }
}
      `},
		{Code: `
class Foo {
  #value: number = 0;

  bar(newValue: { value: number }) {
    ({ value: this.#value } = newValue);
    return this.#value;
  }
}
      `},
		{Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private _name: string;

    public test(value: string) {
      this._name = value;
    }
  };
}
      `},
		{Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    #name: string;

    public test(value: string) {
      this.#name = value;
    }
  };
}
      `},
		{Code: `
class Foo {
  private value: Record<string, number> = {};

  bar(newValue: Record<string, number>) {
    ({ ...this.value } = newValue);
    return this.value;
  }
}
      `},
		{Code: `
class Foo {
  #value: Record<string, number> = {};

  bar(newValue: Record<string, number>) {
    ({ ...this.#value } = newValue);
    return this.#value;
  }
}
      `},
		{Code: `
class Foo {
  private value: number[] = [];

  bar(newValue: number[]) {
    [...this.value] = newValue;
    return this.value;
  }
}
      `},
		{Code: `
class Foo {
  #value: number[] = [];

  bar(newValue: number[]) {
    [...this.#value] = newValue;
    return this.#value;
  }
}
      `},
		{Code: `
class Foo {
  private value: number = 0;

  bar(newValue: number[]) {
    [this.value] = newValue;
    return this.value;
  }
}
      `},
		{Code: `
class Foo {
  #value: number = 0;

  bar(newValue: number[]) {
    [this.#value] = newValue;
    return this.#value;
  }
}
      `},
		{Code: `
        class Test {
          private testObj = {
            prop: '',
          };

          public test(): void {
            this.testObj = '';
          }
        }
      `},
		{Code: `
        class Test {
          #testObj = {
            prop: '',
          };

          public test(): void {
            this.#testObj = '';
          }
        }
      `},
		{Code: `
        class TestObject {
          public prop: number;
        }

        class Test {
          private testObj = new TestObject();

          public test(): void {
            this.testObj = new TestObject();
          }
        }
      `},
		{Code: `
        class TestObject {
          public prop: number;
        }

        class Test {
          #testObj = new TestObject();

          public test(): void {
            this.#testObj = new TestObject();
          }
        }
      `},
		{Code: `
      class TestIntersection {
        private prop: number = 3;

        test() {
          const that = {} as this & { _foo: 'bar' };
          that.prop = 1;
        }
      }
    `},
		{Code: `
      class TestUnion {
        private prop: number = 3;

        test() {
          const that = {} as this | (this & { _foo: 'bar' });
          that.prop = 1;
        }
      }
    `},
		{Code: `
      class TestStaticIntersection {
        private static prop: number;

        test() {
          const that = {} as typeof TestStaticIntersection & { _foo: 'bar' };
          that.prop = 1;
        }
      }
    `},
		{Code: `
      class TestStaticUnion {
        private static prop: number = 1;

        test() {
          const that = {} as
            | typeof TestStaticUnion
            | (typeof TestStaticUnion & { _foo: 'bar' });
          that.prop = 1;
        }
      }
    `},
		{Code: `
      class TestBothIntersection {
        private prop1: number = 1;
        private static prop2: number;

        test() {
          const that = {} as typeof TestBothIntersection & this;
          that.prop1 = 1;
          that.prop2 = 1;
        }
      }
    `},
		{Code: `
      class TestBothIntersection {
        private prop1: number = 1;
        private static prop2: number;

        test() {
          const that = {} as this & typeof TestBothIntersection;
          that.prop1 = 1;
          that.prop2 = 1;
        }
      }
    `},
		{Code: `
      class TestStaticPrivateAccessor {
        private static accessor staticAcc = 1;
      }
    `},
		{Code: `
      class TestStaticPrivateFieldAccessor {
        static accessor #staticAcc = 1;
      }
    `},
		{Code: `
      class TestPrivateAccessor {
        private accessor acc = 3;
      }
    `},
		{Code: `
      class TestPrivateFieldAccessor {
        accessor #acc = 3;
      }
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        class TestIncorrectlyModifiableStatic {
          private static incorrectlyModifiableStatic = 7;
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableStatic {
          private static readonly incorrectlyModifiableStatic = 7;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 53}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableStatic {
          static #incorrectlyModifiableStatic = 7;
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableStatic {
          static readonly #incorrectlyModifiableStatic = 7;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 46}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableStaticArrow {
          private static incorrectlyModifiableStaticArrow = () => 7;
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableStaticArrow {
          private static readonly incorrectlyModifiableStaticArrow = () => 7;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 58}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableStaticArrow {
          static #incorrectlyModifiableStaticArrow = () => 7;
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableStaticArrow {
          static readonly #incorrectlyModifiableStaticArrow = () => 7;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 51}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableInline {
          private incorrectlyModifiableInline = 7;

          public createConfusingChildClass() {
            return class {
              private incorrectlyModifiableInline = 7;
            };
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableInline {
          private readonly incorrectlyModifiableInline = 7;

          public createConfusingChildClass() {
            return class {
              private readonly incorrectlyModifiableInline = 7;
            };
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 46}, {MessageId: `preferReadonly`, Line: 7, Column: 15, EndLine: 7, EndColumn: 50}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableInline {
          #incorrectlyModifiableInline = 7;

          public createConfusingChildClass() {
            return class {
              #incorrectlyModifiableInline = 7;
            };
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableInline {
          readonly #incorrectlyModifiableInline = 7;

          public createConfusingChildClass() {
            return class {
              readonly #incorrectlyModifiableInline = 7;
            };
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 39}, {MessageId: `preferReadonly`, Line: 7, Column: 15, EndLine: 7, EndColumn: 43}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableDelayed {
          private incorrectlyModifiableDelayed = 7;

          public constructor() {
            this.incorrectlyModifiableDelayed = 7;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableDelayed {
          private readonly incorrectlyModifiableDelayed: number = 7;

          public constructor() {
            this.incorrectlyModifiableDelayed = 7;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 47}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableDelayed {
          #incorrectlyModifiableDelayed = 7;

          public constructor() {
            this.#incorrectlyModifiableDelayed = 7;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableDelayed {
          readonly #incorrectlyModifiableDelayed = 7;

          public constructor() {
            this.#incorrectlyModifiableDelayed = 7;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 40}},
		},
		{
			Code: `
        class TestChildClassExpressionModifiable {
          private childClassExpressionModifiable = 7;

          public createConfusingChildClass() {
            return class {
              private childClassExpressionModifiable = 7;

              mutate() {
                this.childClassExpressionModifiable += 1;
              }
            };
          }
        }
      `,
			Output: []string{`
        class TestChildClassExpressionModifiable {
          private readonly childClassExpressionModifiable = 7;

          public createConfusingChildClass() {
            return class {
              private childClassExpressionModifiable = 7;

              mutate() {
                this.childClassExpressionModifiable += 1;
              }
            };
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 49}},
		},
		{
			Code: `
        class TestChildClassExpressionModifiable {
          #childClassExpressionModifiable = 7;

          public createConfusingChildClass() {
            return class {
              #childClassExpressionModifiable = 7;

              mutate() {
                this.#childClassExpressionModifiable += 1;
              }
            };
          }
        }
      `,
			Output: []string{`
        class TestChildClassExpressionModifiable {
          readonly #childClassExpressionModifiable = 7;

          public createConfusingChildClass() {
            return class {
              #childClassExpressionModifiable = 7;

              mutate() {
                this.#childClassExpressionModifiable += 1;
              }
            };
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 42}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePostMinus {
          private incorrectlyModifiablePostMinus = 7;

          public mutate() {
            this.incorrectlyModifiablePostMinus - 1;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePostMinus {
          private readonly incorrectlyModifiablePostMinus = 7;

          public mutate() {
            this.incorrectlyModifiablePostMinus - 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 49}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePostMinus {
          #incorrectlyModifiablePostMinus = 7;

          public mutate() {
            this.#incorrectlyModifiablePostMinus - 1;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePostMinus {
          readonly #incorrectlyModifiablePostMinus = 7;

          public mutate() {
            this.#incorrectlyModifiablePostMinus - 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 42}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePostPlus {
          private incorrectlyModifiablePostPlus = 7;

          public mutate() {
            this.incorrectlyModifiablePostPlus + 1;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePostPlus {
          private readonly incorrectlyModifiablePostPlus = 7;

          public mutate() {
            this.incorrectlyModifiablePostPlus + 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 48}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePostPlus {
          #incorrectlyModifiablePostPlus = 7;

          public mutate() {
            this.#incorrectlyModifiablePostPlus + 1;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePostPlus {
          readonly #incorrectlyModifiablePostPlus = 7;

          public mutate() {
            this.#incorrectlyModifiablePostPlus + 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 41}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePreMinus {
          private incorrectlyModifiablePreMinus = 7;

          public mutate() {
            -this.incorrectlyModifiablePreMinus;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePreMinus {
          private readonly incorrectlyModifiablePreMinus = 7;

          public mutate() {
            -this.incorrectlyModifiablePreMinus;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 48}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePreMinus {
          #incorrectlyModifiablePreMinus = 7;

          public mutate() {
            -this.#incorrectlyModifiablePreMinus;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePreMinus {
          readonly #incorrectlyModifiablePreMinus = 7;

          public mutate() {
            -this.#incorrectlyModifiablePreMinus;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 41}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePrePlus {
          private incorrectlyModifiablePrePlus = 7;

          public mutate() {
            +this.incorrectlyModifiablePrePlus;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePrePlus {
          private readonly incorrectlyModifiablePrePlus = 7;

          public mutate() {
            +this.incorrectlyModifiablePrePlus;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 47}},
		},
		{
			Code: `
        class TestIncorrectlyModifiablePrePlus {
          #incorrectlyModifiablePrePlus = 7;

          public mutate() {
            +this.#incorrectlyModifiablePrePlus;
          }
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiablePrePlus {
          readonly #incorrectlyModifiablePrePlus = 7;

          public mutate() {
            +this.#incorrectlyModifiablePrePlus;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 40}},
		},
		{
			Code: `
        class TestOverlappingClassVariable {
          private overlappingClassVariable = 7;

          public workWithSimilarClass(other: SimilarClass) {
            other.overlappingClassVariable = 7;
          }
        }

        class SimilarClass {
          public overlappingClassVariable = 7;
        }
      `,
			Output: []string{`
        class TestOverlappingClassVariable {
          private readonly overlappingClassVariable = 7;

          public workWithSimilarClass(other: SimilarClass) {
            other.overlappingClassVariable = 7;
          }
        }

        class SimilarClass {
          public overlappingClassVariable = 7;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 43}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableParameter {
          public constructor(private incorrectlyModifiableParameter = 7) {}
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableParameter {
          public constructor(private readonly incorrectlyModifiableParameter = 7) {}
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 30, EndLine: 3, EndColumn: 68}},
		},
		{
			Code: `
        class TestIncorrectlyModifiableParameter {
          public constructor(
            public ignore: boolean,
            private incorrectlyModifiableParameter = 7,
          ) {}
        }
      `,
			Output: []string{`
        class TestIncorrectlyModifiableParameter {
          public constructor(
            public ignore: boolean,
            private readonly incorrectlyModifiableParameter = 7,
          ) {}
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 13, EndLine: 5, EndColumn: 51}},
		},
		{
			Code: `
        class TestCorrectlyNonInlineLambdas {
          private incorrectlyInlineLambda = () => 7;
        }
      `,
			Output: []string{`
        class TestCorrectlyNonInlineLambdas {
          private readonly incorrectlyInlineLambda = () => 7;
        }
      `},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 42}},
			Options: rule_tester.OptionsFromJSON[PreferReadonlyOptions](`{"onlyInlineLambdas":true}`),
		},
		{
			Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private _name: string;
  };
}
      `,
			Output: []string{`
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    private readonly _name: string;
  };
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 4, Column: 5, EndLine: 4, EndColumn: 18}},
		},
		{
			Code: `
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    #name: string;
  };
}
      `,
			Output: []string{`
function ClassWithName<TBase extends new (...args: any[]) => {}>(Base: TBase) {
  return class extends Base {
    readonly #name: string;
  };
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 4, Column: 5, EndLine: 4, EndColumn: 10}},
		},
		{
			Code: `
        class Test {
          private testObj = {
            prop: '',
          };

          public test(): void {
            this.testObj.prop = '';
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {
            prop: '',
          };

          public test(): void {
            this.testObj.prop = '';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {
            prop: '',
          };

          public test(): void {
            this.#testObj.prop = '';
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {
            prop: '',
          };

          public test(): void {
            this.#testObj.prop = '';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class TestObject {
          public prop: number;
        }

        class Test {
          private testObj = new TestObject();

          public test(): void {
            this.testObj.prop = 10;
          }
        }
      `,
			Output: []string{`
        class TestObject {
          public prop: number;
        }

        class Test {
          private readonly testObj = new TestObject();

          public test(): void {
            this.testObj.prop = 10;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 7, Column: 11, EndLine: 7, EndColumn: 26}},
		},
		{
			Code: `
        class TestObject {
          public prop: number;
        }

        class Test {
          #testObj = new TestObject();

          public test(): void {
            this.#testObj.prop = 10;
          }
        }
      `,
			Output: []string{`
        class TestObject {
          public prop: number;
        }

        class Test {
          readonly #testObj = new TestObject();

          public test(): void {
            this.#testObj.prop = 10;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 7, Column: 11, EndLine: 7, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {
            prop: '',
          };
          public test(): void {
            this.testObj.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {
            prop: '',
          };
          public test(): void {
            this.testObj.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {
            prop: '',
          };
          public test(): void {
            this.#testObj.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {
            prop: '',
          };
          public test(): void {
            this.#testObj.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj!.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj!.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj!.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj!.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj.prop.prop = '';
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj.prop.prop = '';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj.prop.prop = '';
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj.prop.prop = '';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj.prop.doesSomething();
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj.prop.doesSomething();
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj.prop.doesSomething();
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj.prop.doesSomething();
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj?.prop.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj?.prop.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj?.prop.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj?.prop.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj?.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj?.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj?.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj?.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private testObj = {};
          public test(): void {
            this.testObj!.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly testObj = {};
          public test(): void {
            this.testObj!.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 26}},
		},
		{
			Code: `
        class Test {
          #testObj = {};
          public test(): void {
            this.#testObj!.prop?.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          readonly #testObj = {};
          public test(): void {
            this.#testObj!.prop?.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 19}},
		},
		{
			Code: `
        class Test {
          private prop: number = 3;

          test() {
            const that = {} as this & { _foo: 'bar' };
            that._foo = 1;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: number = 3;

          test() {
            const that = {} as this & { _foo: 'bar' };
            that._foo = 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop: number = 3;

          test() {
            const that = {} as this | (this & { _foo: 'bar' });
            that.prop;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: number = 3;

          test() {
            const that = {} as this | (this & { _foo: 'bar' });
            that.prop;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop: number;

          constructor() {
            const that = {} as this & { _foo: 'bar' };
            that.prop = 1;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: number;

          constructor() {
            const that = {} as this & { _foo: 'bar' };
            that.prop = 1;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = 'hello';

          constructor() {
            this.prop = 'world';
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: string = 'hello';

          constructor() {
            this.prop = 'world';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = 'hello';
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = 'hello';
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        declare const hello: 'hello';

        class Test {
          private prop = hello;

          constructor() {
            this.prop = 'world';
          }
        }
      `,
			Output: []string{`
        declare const hello: 'hello';

        class Test {
          private readonly prop = hello;

          constructor() {
            this.prop = 'world';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 11, EndLine: 5, EndColumn: 23}},
		},
		{
			Code: `
        declare const hello: 'hello';

        class Test {
          private prop = hello;
        }
      `,
			Output: []string{`
        declare const hello: 'hello';

        class Test {
          private readonly prop = hello;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 11, EndLine: 5, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = 10;

          constructor() {
            this.prop = 11;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: number = 10;

          constructor() {
            this.prop = 11;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = 10;
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = 10;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        declare const hello: 10;

        class Test {
          private prop = hello;

          constructor() {
            this.prop = 11;
          }
        }
      `,
			Output: []string{`
        declare const hello: 10;

        class Test {
          private readonly prop = hello;

          constructor() {
            this.prop = 11;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 11, EndLine: 5, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = true;

          constructor() {
            this.prop = false;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: boolean = true;

          constructor() {
            this.prop = false;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = true;
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = true;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        declare const hello: true;

        class Test {
          private prop = hello;

          constructor() {
            this.prop = false;
          }
        }
      `,
			Output: []string{`
        declare const hello: true;

        class Test {
          private readonly prop = hello;

          constructor() {
            this.prop = false;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 11, EndLine: 5, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        class Test {
          private prop = Foo.Bar;

          constructor() {
            this.prop = Foo.Bazz;
          }
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        class Test {
          private readonly prop: Foo = Foo.Bar;

          constructor() {
            this.prop = Foo.Bazz;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 8, Column: 11, EndLine: 8, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        class Test {
          private prop = Foo.Bar;
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        class Test {
          private readonly prop = Foo.Bar;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 8, Column: 11, EndLine: 8, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        const foo = Foo.Bar;

        class Test {
          private prop = foo;

          constructor() {
            this.prop = foo;
          }
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        const foo = Foo.Bar;

        class Test {
          private readonly prop: Foo = foo;

          constructor() {
            this.prop = foo;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 10, Column: 11, EndLine: 10, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        const foo = Foo.Bar;

        class Test {
          private prop = foo;
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        const foo = Foo.Bar;

        class Test {
          private readonly prop = foo;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 10, Column: 11, EndLine: 10, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        declare const foo: Foo;

        class Test {
          private prop = foo;
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        declare const foo: Foo;

        class Test {
          private readonly prop = foo;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 10, Column: 11, EndLine: 10, EndColumn: 23}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        const bar = Foo.Bar;

        function wrapper() {
          const Foo = 10;

          class Test {
            private prop = bar;

            constructor() {
              this.prop = bar;
            }
          }
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        const bar = Foo.Bar;

        function wrapper() {
          const Foo = 10;

          class Test {
            private readonly prop = bar;

            constructor() {
              this.prop = bar;
            }
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 13, Column: 13, EndLine: 13, EndColumn: 25}},
		},
		{
			Code: `
        enum Foo {
          Bar,
          Bazz,
        }

        const bar = Foo.Bar;

        function wrapper() {
          type Foo = 10;

          class Test {
            private prop = bar;

            constructor() {
              this.prop = bar;
            }
          }
        }
      `,
			Output: []string{`
        enum Foo {
          Bar,
          Bazz,
        }

        const bar = Foo.Bar;

        function wrapper() {
          type Foo = 10;

          class Test {
            private readonly prop = bar;

            constructor() {
              this.prop = bar;
            }
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 13, Column: 13, EndLine: 13, EndColumn: 25}},
		},
		{
			Code: `
        const Bar = (function () {
          enum Foo {
            Bar,
            Bazz,
          }

          return Foo;
        })();

        const bar = Bar.Bar;

        class Test {
          private prop = bar;

          constructor() {
            this.prop = bar;
          }
        }
      `,
			Output: []string{`
        const Bar = (function () {
          enum Foo {
            Bar,
            Bazz,
          }

          return Foo;
        })();

        const bar = Bar.Bar;

        class Test {
          private readonly prop = bar;

          constructor() {
            this.prop = bar;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 14, Column: 11, EndLine: 14, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = { foo: 'bar' };
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = { foo: 'bar' };
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = { foo: 'bar' };

          constructor() {
            this.prop = { foo: 'bazz' };
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = { foo: 'bar' };

          constructor() {
            this.prop = { foo: 'bazz' };
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = [1, 2, 'three'];
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = [1, 2, 'three'];
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = [1, 2, 'three'];

          constructor() {
            this.prop = [1, 2, 'four'];
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = [1, 2, 'three'];

          constructor() {
            this.prop = [1, 2, 'four'];
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class X {
          private _isValid = true;

          getIsValid = () => this._isValid;

          constructor(data?: {}) {
            if (!data) {
              this._isValid = false;
            }
          }
        }
      `,
			Output: []string{`
        class X {
          private readonly _isValid: boolean = true;

          getIsValid = () => this._isValid;

          constructor(data?: {}) {
            if (!data) {
              this._isValid = false;
            }
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 27}},
		},
		{
			Code: `
        class Test {
          private prop: string = 'hello';
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: string = 'hello';
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop: string | number = 'hello';
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: string | number = 'hello';
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop: string;

          constructor() {
            this.prop = 'hello';
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop: string;

          constructor() {
            this.prop = 'hello';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop;

          constructor() {
            this.prop = 'hello';
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop;

          constructor() {
            this.prop = 'hello';
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop;

          constructor(x: boolean) {
            if (x) {
              this.prop = 'hello';
            } else {
              this.prop = 10;
            }
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop;

          constructor(x: boolean) {
            if (x) {
              this.prop = 'hello';
            } else {
              this.prop = 10;
            }
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        declare const hello: 'hello' | 10;

        class Test {
          private prop = hello;

          constructor() {
            this.prop = 10;
          }
        }
      `,
			Output: []string{`
        declare const hello: 'hello' | 10;

        class Test {
          private readonly prop = hello;

          constructor() {
            this.prop = 10;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 5, Column: 11, EndLine: 5, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = null;
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = null;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = null;

          constructor() {
            this.prop = null;
          }
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = null;

          constructor() {
            this.prop = null;
          }
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = 'hello' as string;
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = 'hello' as string;
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
		{
			Code: `
        class Test {
          private prop = Promise.resolve('hello');
        }
      `,
			Output: []string{`
        class Test {
          private readonly prop = Promise.resolve('hello');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: `preferReadonly`, Line: 3, Column: 11, EndLine: 3, EndColumn: 23}},
		},
	})
}
