package related_getter_setter_pairs

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestRelatedGetterSetterPairs(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &RelatedGetterSetterPairsRule, []rule_tester.ValidTestCase{
		{Code: `
interface Example {
  get value(): string;
  set value(newValue: string);
}
    `},
		{Code: `
interface Example {
  get value(): string | undefined;
  set value();
}
    `},
		{Code: `
interface Example {
  get value(): string | undefined;
  set value(newValue: string, invalid: string);
}
    `},
		{Code: `
interface Example {
  get value(): string;
  set value(newValue: string | undefined);
}
    `},
		{Code: `
interface Example {
  get value(): number;
}
    `},
		{Code: `
interface Example {
  get value(): number;
  set value();
}
    `},
		{Code: `
interface Example {
  set value(newValue: string);
}
    `},
		{Code: `
interface Example {
  set value();
}
    `},
		{Code: `
type Example = {
  get value();
};
    `},
		{Code: `
type Example = {
  set value();
};
    `},
		{Code: `
class Example {
  get value() {
    return '';
  }
}
    `},
		{Code: `
class Example {
  get value() {
    return '';
  }
  set value() {}
}
    `},
		{Code: `
class Example {
  get value() {
    return '';
  }
  set value(param) {}
}
    `},
		{Code: `
class Example {
  get value() {
    return '';
  }
  set value(param: number) {}
}
    `},
		{Code: `
class Example {
  set value() {}
}
    `},
		{Code: `
type Example = {
  get value(): number;
  set value(newValue: number);
};
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
interface Example {
  get value(): string | undefined;
  set value(newValue: string);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
interface Example {
  get value(): number;
  set value(newValue: string);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
type Example = {
  get value(): number;
  set value(newValue: string);
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 22,
				},
			},
		},
		{
			Code: `
class Example {
  get value(): boolean {
    return true;
  }
  set value(newValue: string) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
type GetType = { a: string; b: string };

declare class Foo {
  get a(): GetType;

  set a(x: { c: string });
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      5,
					Column:    12,
					EndLine:   5,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
type GetType = { a: string; b: string };

type SetTypeUnused = { c: string };

declare class Foo {
  get a(): GetType;

  set a(x: { c: string });
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      7,
					Column:    12,
					EndLine:   7,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
type GetType = { a: string; b: string };

type SetType = { c: string };

declare class Foo {
  get a(): GetType;

  set a(x: SetType);
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "mismatch",
					Line:      7,
					Column:    12,
					EndLine:   7,
					EndColumn: 19,
				},
			},
		},
	})
}
