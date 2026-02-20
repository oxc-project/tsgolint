package no_useless_coalescing

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoUselessCoalescingRule(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUselessCoalescingRule, []rule_tester.ValidTestCase{
		{
			Code: `
const x: string = getString();
x || undefined;
      `,
		},
		{
			Code: `
const x: string | undefined = getMaybeString();
x || '';
      `,
		},
		{
			Code: `
const x: string | null | undefined = getMaybeString();
x ?? undefined;
      `,
		},
		{
			Code: `
declare const x: string;
x || void sideEffect();
      `,
		},
		{
			Code: `
const x: string = getString();
x || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
		},
		{
			Code: `
const x: string | null = getMaybeString();
x ?? undefined;
      `,
		},
		{
			Code: `
declare const x: string[] | null;
x || undefined;
      `,
		},
		{
			Code: `
const x: number | undefined = getMaybeNumber();
x || undefined;
      `,
		},
		{
			Code: `
declare const maybeObj: { name: string } | undefined;
maybeObj || { name: 'fallback' };
      `,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
const x: string = getString();
x || '';
      `,
			Output: []string{`
const x: string = getString();
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
const x: boolean = getBoolean();
x || false;
      `,
			Output: []string{`
const x: boolean = getBoolean();
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
const x: bigint = getBigInt();
x || 0n;
      `,
			Output: []string{`
const x: bigint = getBigInt();
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
const myVar = 'hello';
myVar || undefined;
      `,
			Output: []string{`
const myVar = 'hello';
myVar;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
declare const myLiteral: 'hello';
myLiteral || undefined;
      `,
			Output: []string{`
declare const myLiteral: 'hello';
myLiteral;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
declare const maybeLiteral: 'hello' | undefined;
maybeLiteral || undefined;
      `,
			Output: []string{`
declare const maybeLiteral: 'hello' | undefined;
maybeLiteral;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const obj: { name: string };
obj || undefined;
      `,
			Output: []string{`
declare const obj: { name: string };
obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
declare const obj: { name: string };
obj || { name: 'fallback' };
      `,
			Output: []string{`
declare const obj: { name: string };
obj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
declare const maybeObj: { name: string } | undefined;
maybeObj || undefined;
      `,
			Output: []string{`
declare const maybeObj: { name: string } | undefined;
maybeObj;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const x: string[] | undefined;
x || undefined;
      `,
			Output: []string{`
declare const x: string[] | undefined;
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const x: string[] | undefined;
x || void 0;
      `,
			Output: []string{`
declare const x: string[] | undefined;
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
const x: string = getString();
x ?? 'fallback';
      `,
			Output: []string{`
const x: string = getString();
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "uselessCoalescing"}},
		},
		{
			Code: `
declare const x: string | undefined;
x ?? undefined;
      `,
			Output: []string{`
declare const x: string | undefined;
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const x: string | undefined;
x ?? void 0;
      `,
			Output: []string{`
declare const x: string | undefined;
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const x: true | undefined;
x || undefined;
      `,
			Output: []string{`
declare const x: true | undefined;
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
declare const maybeArray: string[] | undefined;
const sample10 = maybeArray || undefined || undefined || undefined;
      `,
			Output: []string{`
declare const maybeArray: string[] | undefined;
const sample10 = maybeArray;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "redundantUndefinedFallback"},
				{MessageId: "redundantUndefinedFallback"},
				{MessageId: "redundantUndefinedFallback"},
			},
		},
		{
			Code: `
declare const x: string[] | undefined;
(x || undefined);
      `,
			Output: []string{`
declare const x: string[] | undefined;
(x);
      `},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "redundantUndefinedFallback"}},
		},
		{
			Code: `
const x: string = getString();
x || '';
      `,
			TSConfig: "tsconfig.unstrict.json",
			Output: []string{`
const x: string = getString();
x;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noStrictNullCheck"},
				{MessageId: "uselessCoalescing"},
			},
		},
		{
			Code: `
declare const x: string | undefined;
x || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "falsyUndefinedNormalization"}},
		},
		{
			Code: `
declare const x: number | undefined;
x || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "falsyUndefinedNormalization"}},
		},
		{
			Code: `
declare const x: boolean | undefined;
x || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "falsyUndefinedNormalization"}},
		},
		{
			Code: `
interface User {
  name?: string;
}

declare const user: User | undefined;
user?.name || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "falsyUndefinedNormalization"}},
		},
		{
			Code: `
declare const x: string | null | undefined;
x || undefined;
      `,
			Options: rule_tester.OptionsFromJSON[NoUselessCoalescingOptions](`{"detectFalsyValues": true}`),
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "falsyUndefinedNormalization"}},
		},
	})
}
