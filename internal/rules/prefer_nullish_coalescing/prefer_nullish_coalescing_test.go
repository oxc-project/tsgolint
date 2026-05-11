package prefer_nullish_coalescing

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func nullishSuggestion(output string) []rule_tester.InvalidTestCaseSuggestion {
	return []rule_tester.InvalidTestCaseSuggestion{{
		MessageId: "suggestNullishCoalescing",
		Output:    output,
	}}
}

func TestPreferNullishCoalescingRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule, []rule_tester.ValidTestCase{
		{Code: `x !== undefined && x !== null ? x : y;`, Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: true}},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `, Options: PreferNullishCoalescingOptions{IgnoreIfStatements: true}},
		{Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) foo = makeFoo();
}
      `, Options: PreferNullishCoalescingOptions{IgnoreIfStatements: true}},
		{Code: `
      declare let x: never;
      declare let y: number;
      x || y;
    `},
		{Code: `
      declare let x: never;
      declare let y: number;
      x ? x : y;
    `},
		{Code: `
      declare let x: never;
      declare let y: number;
      !x ? y : x;
    `},
		{Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `},
		{Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== null ? defaultBoxOptional.a?.b : getFallbackBox();
    `},
		{Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
    `},
		{Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | null } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
    `},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`)},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": false }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`)},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`)},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": true }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": false }}`)},
		{Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true, "string": false }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "string": true }}`)},
		{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a || b);
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || b || c);
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a || (b && c));
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a || b) ?? c);
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ?? (b || c));
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? b || c : 'fail');
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(a ? 'success' : b || c);
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(((a = b), b || c));
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean((a ? a : b) || c);
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

const test = Boolean(c || (!a ? b : a));
      `, Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || b || c) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a || (b && c)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a || b) ?? c) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ?? (b || c)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? b || c : 'fail') {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (a ? 'success' : b || c) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (((a = b), b || c)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | undefined;
let b: string | undefined;

if (!(a || b)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | undefined;
let b: string | undefined;

if (!!(a || b)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

if (a ? a : b) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

if (!a ? b : a) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if ((a ? a : b) || c) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;
let c: string | boolean | undefined;

if (c || (!a ? b : a)) {
}
      `, Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true}},
		{Code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `, Options: PreferNullishCoalescingOptions{IgnorePrimitives: utils.BoolOrValue[IgnorePrimitivesOptions](true)}},
		{Code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "number": true }}`)},
		{Code: `
declare const a: unknown;
const b = a || 'bar';
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": false, "string": false }}`)},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `this != undefined ? this : y;`,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`this ?? y;`),
				},
			},
		},
		{
			Code: `
declare let x: string[] | null;
if (x) {
}
      `,
			TSConfig: "tsconfig.unstrict.json",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noStrictNullCheck",
				},
			},
		},
		{
			Code: `
declare let x: string | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: string | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: number | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: boolean | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: boolean | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: bigint | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: bigint | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: string | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: string | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: number | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: boolean | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: bigint | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: '' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: '' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: ` + "`" + "`" + ` | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: ` + "`" + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 0 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 0n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: false | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: false | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: '' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: '' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: ` + "`" + "`" + ` | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: ` + "`" + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: false | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: false | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: true | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: ` + "`" + `hello${'string'}` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: true | false | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		// Template literal types are handled as string-like types
		{
			Code: `
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": true, "string": false }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 'a' | ` + "`" + `b` + "`" + ` | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | false | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | false | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: true | false | null | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": false, "boolean": true, "number": false, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | false | null | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": false, "number": true, "string": true }}`),

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: true | false | null | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: null;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let x: null;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
const x = undefined;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
const x = undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
null || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
null ?? y;
      `),
				},
			},
		},
		{
			Code: `
undefined || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
undefined ?? y;
      `),
				},
			},
		},
		{
			Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a || b);
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a ?? b);
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a || b);
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a ?? b);
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a || b);
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a ?? b);
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a || b;
});
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a ?? b;
});
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a || b));
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a ?? b));
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a || b));
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a ?? b));
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ? a : b);
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ?? b);
      `),
				},
			},
		},
		{
			Code: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(!a ? b : a);
      `,
			Options: PreferNullishCoalescingOptions{IgnoreBooleanCoercion: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(a ?? b);
      `),
				},
			},
		},
		{
			Code: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a || b)) {
}
      `,
			Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a ?? b)) {
}
      `),
				},
			},
		},
		{
			Code: `
declare const a: string | undefined;
declare const b: string;

if (+(a || b)) {
}
      `,
			Options: PreferNullishCoalescingOptions{IgnoreConditionalTests: true},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const a: string | undefined;
declare const b: string;

if (+(a ?? b)) {
}
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox || getFallbackBox();
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ? defaultBox : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
declare const x: any;
declare const y: any;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const x: any;
declare const y: any;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare const x: unknown;
declare const y: any;
x || y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const x: unknown;
declare const y: any;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b != null ? defaultBoxOptional.a.b : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a?.b : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ? defaultBoxOptional.a.b : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a?.b
  : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b !== undefined && defaultBoxOptional.a?.b !== null
  ? defaultBoxOptional.a.b
  : getFallbackBox();
      `,
			Options: PreferNullishCoalescingOptions{IgnoreTernaryTests: false},

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `),
				},
			},
		},
		{
			Code: `
declare let x: unknown;
declare let y: number;
!x ? y : x;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: unknown;
declare let y: number;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: unknown;
declare let y: number;
x ? x : y;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: unknown;
declare let y: number;
x ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: { n: unknown };
!x.n ? y : x.n;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: { n: unknown };
x.n ?? y;
      `),
				},
			},
		},
		{
			Code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x['a'] : 'foo';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `),
				},
			},
		},
		// Rule handles mixed property access syntax (bracket vs dot notation)
		{
			Code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x.a : 'foo';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `),
				},
			},
		},
		// Rule handles mixed property access syntax (bracket vs dot notation)
		{
			Code: `
declare let x: { a: string } | null;

x?.a != null ? x['a'] : 'foo';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare let x: { a: string } | null;

x?.a ?? 'foo';
      `),
				},
			},
		},
		{
			Code: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] != null ? x[a] : 'foo';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] ?? 'foo';
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (!foo) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ||= makeFoo();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === null) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo = makeFoo();
  const bar = 42;
  return bar;
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ||= makeFoo();
  const bar = 42;
  return bar;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `),
				},
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo === undefined || foo === null) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo.a == null) {
    foo.a = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo?.a == null) {
    foo.a = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `),
				},
			},
		},
		{
			Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
  if (foo == null) {
    foo = makeFoo();
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
  foo ??= makeFoo();
}
      `),
				},
			},
		},
		// TODO: Skip - rule handles comments differently than expected
		{
			Skip: true,
			Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: string | null;
declare function makeFoo(): string;

// comment before 1
/* comment before 2 */
/* comment before 3
    which is multiline
  */
/**
   * comment before 4
   * which is also multiline
   */
foo ??= makeFoo(); // comment inline
// comment after 1
/* comment after 2 */
/* comment after 3
    which is multiline
  */
/**
   * comment after 4
   * which is also multiline
   */
      `),
				},
			},
		},
		// TODO: Skip - rule handles comments differently than expected
		{
			Skip: true,
			Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) {
  // comment before 1
  /* comment before 2 */
  /* comment before 3
    which is multiline
  */
  /**
   * comment before 4
   * which is also multiline
   */
  foo = makeFoo(); // comment inline
  // comment after 1
  /* comment after 2 */
  /* comment after 3
    which is multiline
  */
  /**
   * comment after 4
   * which is also multiline
   */
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: string | null;
declare function makeFoo(): string;

// comment before 1
/* comment before 2 */
/* comment before 3
    which is multiline
  */
/**
   * comment before 4
   * which is also multiline
   */
foo ??= makeFoo(); // comment inline
// comment after 1
/* comment after 2 */
/* comment after 3
    which is multiline
  */
/**
   * comment after 4
   * which is also multiline
   */
      `),
				},
			},
		},
		{
			Code: `
declare let foo: string | null;
declare function makeFoo(): string;

if (foo == null) /* comment before 1 */ /* comment before 2 */ foo = makeFoo(); // comment inline
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
      `),
				},
			},
		},
		{
			Code: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  if (((((foo.a)) == null))) {
    ((((((((foo).a))))) = makeString()));
  }
}
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: nullishSuggestion(`
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
      `),
				},
			},
		},
		{
			Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : b ? 1 : 2;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `),
				},
			},
		},
		{
			Code: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ? a : (b ? 1 : 2);
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `),
				},
			},
		},
		{
			Code: `
declare const c: string | null;
c !== null ? c : c ? 1 : 2;
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: nullishSuggestion(`
declare const c: string | null;
c ?? (c ? 1 : 2);
      `),
				},
			},
		},
		// https://github.com/oxc-project/tsgolint/issues/604
		// Test for parenthesized logical expressions
		{
			Code: `
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = (a && b) || c || 'd';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = ((a && b) ?? c) || 'd';
      `),
				},
			},
		},
		// https://github.com/oxc-project/tsgolint/issues/604
		// Test for deeply nested parenthesized logical expressions
		{
			Code: `
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = ((a && b)) || c || 'd';
      `,

			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = (((a && b)) ?? c) || 'd';
      `),
				},
			},
		},
		// https://github.com/oxc-project/tsgolint/issues/604
		// Test for non-parenthesized logical expression
		{
			Code: `
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = a && b || c || 'd';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare let a: string | null;
declare let b: string;
declare let c: string;
const x = a && (b ?? c) || 'd';
      `),
				},
			},
		},
		// https://github.com/oxc-project/oxc/issues/21978
		{
			Code: `
declare const isLoading: boolean;
declare const isPending: boolean | undefined;
declare const fallback: boolean;
export const result = isLoading || isPending || fallback;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const isLoading: boolean;
declare const isPending: boolean | undefined;
declare const fallback: boolean;
export const result = (isLoading || isPending) ?? fallback;
      `),
				},
			},
		},
		{
			Code: `
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
export const result = value || fallback && alternate;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
export const result = value ?? (fallback && alternate);
      `),
				},
			},
		},
		{
			Code: `
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
export const result = value || (fallback && alternate);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
export const result = value ?? (fallback && alternate);
      `),
				},
			},
		},
		{
			Code: `
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
declare const finalFallback: string;
export const result = value || fallback && alternate || finalFallback;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: nullishSuggestion(`
declare const value: string | undefined;
declare const fallback: string;
declare const alternate: string;
declare const finalFallback: string;
export const result = (value ?? (fallback && alternate)) || finalFallback;
      `),
				},
			},
		},
	})
}
