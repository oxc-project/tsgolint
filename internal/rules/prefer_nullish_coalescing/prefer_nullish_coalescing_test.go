package prefer_nullish_coalescing

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferNullishCoalescingRule(t *testing.T) {
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
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": true, "boolean": true, "number": false, "string": true }}`) },
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x || y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare let x: 0 | 'foo' | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ? x : y;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
!x ? y : x;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
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
      `, Options: PreferNullishCoalescingOptions{IgnorePrimitives: true}},
		{Code: `
declare const a: any;
declare const b: any;
a ? a : b;
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
		{Code: `
declare const a: unknown;
const b = a || 'bar';
      `, Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) },
	}, []rule_tester.InvalidTestCase{
		{
			Code: `this != undefined ? this : y;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							Output: `this ?? y;`,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: string[] | null;
if (x) {
}
      `,
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
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: string | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: number | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: boolean | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: boolean | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: bigint | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: bigint | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: string | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: string | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: number | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: number | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: boolean | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: boolean | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: bigint | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: bigint | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: '' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: '' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: \\`\\` | undefined;\nx || y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: \\`\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: false | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: false | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: '' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: '' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: \\`\\` | undefined;\nx ? x : y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: \\`\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: false | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: false | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\nx || y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\nx ? x : y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\n!x ? y : x;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: \\`hello\\${'string'}\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: 'a' | \\`b\\` | undefined;\nx || y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: 'a' | \\`b\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 'a' | 'b' | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 'a' | 'b' | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: 'a' | \\`b\\` | undefined;\nx ? x : y;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: 'a' | \\`b\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: "\ndeclare let x: 'a' | \\`b\\` | undefined;\n!x ? y : x;\n      ",
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: "\ndeclare let x: 'a' | \\`b\\` | undefined;\nx ?? y;\n      ",
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1 | 2 | 3 | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1 | 2 | 3 | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0n | 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 1n | 2n | 3n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 1n | 2n | 3n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
x || y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: 0 | 1 | 0n | 1n | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: 0 | 1 | 0n | 1n | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
x ? x : y;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: true | false | null | undefined;
!x ? y : x;
      `,
			Options: rule_tester.OptionsFromJSON[PreferNullishCoalescingOptions](`{"ignorePrimitives": { "bigint": XX, "boolean": XX, "number": XX, "string": XX }}`) ,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: true | false | null | undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: null;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
const x = undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
null ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
undefined ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum | undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
enum Enum {
  A = 0,
  B = 1,
  C = 2,
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum | undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
enum Enum {
  A = 'a',
  B = 'b',
  C = 'c',
}
declare let x: Enum.A | Enum.B | undefined;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;
let c: boolean | undefined;

const x = Boolean(a ?? b);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = String(a ?? b);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(() => a ?? b);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(function weird() {
  return a ?? b;
});
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

const x = Boolean(f(a ?? b));
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(1 + (a ?? b));
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

const x = Boolean(a ?? b);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | boolean | undefined;
let b: string | boolean | undefined;

const test = Boolean(a ?? b);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | true | undefined;
let b: string | boolean | undefined;

declare function f(x: unknown): unknown;

if (f(a ?? b)) {
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare const a: string | undefined;
declare const b: string;

if (+(a ?? b)) {
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBox: Box | undefined;

defaultBox ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare const x: any;
declare const y: any;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare const x: unknown;
declare const y: any;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
interface Box {
  value: string;
}
declare function getFallbackBox(): Box;
declare const defaultBoxOptional: { a?: { b?: Box | undefined } };

defaultBoxOptional.a?.b ?? getFallbackBox();
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: unknown;
declare let y: number;
x ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: { n: unknown };
x.n ?? y;
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: { a: string } | null;

x?.['a'] != null ? x.a : 'foo';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: { a: string } | null;

x?.['a'] ?? 'foo';
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let x: { a: string } | null;

x?.a != null ? x['a'] : 'foo';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverTernary",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let x: { a: string } | null;

x?.a ?? 'foo';
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
const a = 'b';
declare let x: { a: string; b: string } | null;

x?.[a] ?? 'foo';
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
				},
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) {
    foo ??= makeFoo();
  }
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
						},
					},
				},
				{
					MessageId: "preferNullishOverOr",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  if (foo == null) foo ??= makeFoo();
  const bar = 42;
  return bar;
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null | undefined;
declare function makeFoo(): { a: string };

function lazyInitialize() {
  foo ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string } | null;
declare function makeFoo(): string;

function lazyInitialize() {
  foo.a ??= makeFoo();
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  if (foo == null) {
    // comment
    foo = makeFoo();
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferNullishOverAssignment",
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: string | null;
declare function makeFoo(): string;

function lazyInitialize() {
  // comment
foo ??= makeFoo();
}
      `,
						},
					},
				},
			},
		},
		{
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
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
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: string | null;
declare function makeFoo(): string;

/* comment before 1 */ /* comment before 2 */ foo ??= makeFoo(); // comment inline
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare let foo: { a: string | null };
declare function makeString(): string;

function weirdParens() {
  ((foo).a) ??= makeString();
}
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
let a: string | undefined;
let b: { message: string } | undefined;

const foo = a ?? (b ? 1 : 2);
      `,
						},
					},
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
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestNullish",
							Output: `
declare const c: string | null;
c ?? (c ? 1 : 2);
      `,
						},
					},
				},
			},
		},
	})
}
