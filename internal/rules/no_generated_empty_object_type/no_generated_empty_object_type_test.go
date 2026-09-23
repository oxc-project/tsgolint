package no_generated_empty_object_type

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// Ported from typescript-eslint at 370cbe57d (PR #12854).
func TestNoGeneratedEmptyObjectTypeRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoGeneratedEmptyObjectTypeRule, []rule_tester.ValidTestCase{
		{Code: `type Constructor = new () => object; type Alias = Exclude<Constructor, undefined>;`},
		{Code: `type Data = { a: string }; type Result = (Omit<Data, 'a'>) & { b: number };`},

		{Code: `
type Data = { name: string; num: number };
type Expected = Omit<Data, 'name'>;
    `},
		{Code: `
type Data = { name: string; num: number };
declare function doSomething<T extends Omit<Data, 'name'>>(param: T): void;
    `},
		{Code: `
type Names = Array<string>;
    `},
		{Code: `
type Empty = {};
    `},
		{Code: `
type Explicit = {} | { value: number };
    `},
		{Code: `
interface Empty {}
type Alias = Empty;
    `},
		{Code: `
class Empty {}
type Alias = Empty;
    `},
		{Code: `
type Dictionary = Record<string, never>;
    `},
		{Code: `
type Callable = () => void;
type Aliased = Exclude<Callable, undefined>;
    `},
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
type Intersected = Omit<NullableData, 'name'> & { other: string };
    `},
		{Code: `
type Data = { name: string; value: number };
type Data2 = { name: string };
type DistributiveOmit<T, K extends PropertyKey> = T extends unknown
  ? Omit<T, K>
  : never;
type Expected = DistributiveOmit<Data | Data2, 'name'> & {
  other: string;
};
    `},
		{Code: `
declare function getEnumNames<T extends string>(
  myEnum: Record<T, unknown>,
): T[];
    `},
		{Code: `
type MakeRequired<Base, Key extends keyof Base> = Omit<Base, Key> &
  Required<Record<Key, NonNullable<Base[Key]>>>;
    `},
		{Code: `
type Emptied<T> = Omit<T, keyof T>;
    `},
		{Code: `
type Keys<T> = T extends infer U ? keyof U : never;
type Mapped<T extends object> = { [Key in Keys<T>]: Key };
type Referenced<T extends object> = Mapped<T>;
    `},
		{Code: `
type Keys<T> = T extends infer U ? keyof U : never;
type Mapped<T extends object> = { [Key in Keys<T>]: Key };
type Indexed<T extends object> = Mapped<T>[Keys<T>];
    `},
		{Code: `
type Keys<T> = T extends infer U ? keyof U : never;
type Remapped<T extends object> = {
  [Key in Keys<T> as ` + "`" + `get${string & Key}` + "`" + `]: () => void;
};
type Referenced<T extends object> = Remapped<T>;
    `},
		{Code: `
type Keys<T> = T extends infer U ? keyof U : never;
type ReadonlyMapped<T extends object> = { readonly [Key in Keys<T>]: number };
type Referenced<T extends object> = ReadonlyMapped<T>;
    `},
	}, []rule_tester.InvalidTestCase{
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
type Unexpected = Omit<NullableData, 'name'>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 45, EndLine: 4, Line: 4}}},
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
function doSomething<T extends Omit<NullableData, 'name'>>(param: T) {}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 32, EndColumn: 58, EndLine: 4, Line: 4}}},
		{Code: `
type Data = { name: string; num: number };
type Unexpected = Pick<Data, never>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 36, EndLine: 3, Line: 3}}},
		{Code: `
type Unexpected = NonNullable<unknown>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 39, EndLine: 2, Line: 2}}},
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
declare const unexpected: Omit<NullableData, 'name'>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 27, EndColumn: 53, EndLine: 4, Line: 4}}},
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
type Unexpected = Omit<NullableData, 'name'>;
type Referencing = Unexpected;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 45, EndLine: 4, Line: 4}}},
		{Code: `
type Data = { name: string; num: number };
type NullableData = null | Data;
type Unexpected = Array<Omit<NullableData, 'name'>>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 25, EndColumn: 51, EndLine: 4, Line: 4}}},
		{Code: `
type Data = { name: string; value: number };
type Data2 = { name: string };
type DistributiveOmit<T, K extends PropertyKey> = T extends unknown
  ? Omit<T, K>
  : never;
type Unexpected = DistributiveOmit<Data | Data2, 'name'>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 57, EndLine: 7, Line: 7}}},
		{Code: `
type Data = { a: string };
type Unexpected = Omit<Data, 'a'> & Omit<Data, 'a'>;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noGeneratedEmptyObjectType", Column: 19, EndColumn: 52, EndLine: 3, Line: 3}}},
	})
}
