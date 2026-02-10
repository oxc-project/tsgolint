package prefer_regexp_exec

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferRegexpExecRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &PreferRegexpExecRule, []rule_tester.ValidTestCase{
		{Code: `'something'.match();`},
		{Code: `'something'.match(/thing/g);`},
		{Code: `
const text = 'something';
const search = /thing/g;
text.match(search);
    `},
		{Code: `
const match = (s: RegExp) => 'something';
match(/thing/);
    `},
		{Code: `
const a = { match: (s: RegExp) => 'something' };
a.match(/thing/);
    `},
		{Code: `
function f(s: string | string[]) {
  s.match(/e/);
}
    `},
		{Code: `(Math.random() > 0.5 ? 'abc' : 123).match(2);`},
		{Code: `'212'.match(2);`},
		{Code: `'212'.match(+2);`},
		{Code: `'oNaNo'.match(NaN);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(Infinity);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(+Infinity);`},
		{Code: `'Infinity contains -Infinity and +Infinity in JavaScript.'.match(-Infinity);`},
		{Code: `'void and null'.match(null);`},
		{Code: `
const matchers = ['package-lock.json', /regexp/];
const file = '';
matchers.some(matcher => !!file.match(matcher));
    `},
		{Code: `
const matchers = [/regexp/, 'package-lock.json'];
const file = '';
matchers.some(matcher => !!file.match(matcher));
    `},
		{Code: `
const matchers = [{ match: (s: RegExp) => false }];
const file = '';
matchers.some(matcher => !!file.match(matcher));
    `},
		{Code: `
function test(pattern: string) {
  'hello hello'.match(RegExp(pattern, 'g'))?.reduce(() => []);
}
    `},
		{Code: `
function test(pattern: string) {
  'hello hello'.match(new RegExp(pattern, 'gi'))?.reduce(() => []);
}
    `},
		{Code: `
function test(text: string, pattern: string, flags: string) {
  text.match(new RegExp(pattern, flags));
}
    `},
		{Code: `
const matchCount = (str: string, re: RegExp) => {
  return (str.match(re) || []).length;
};
    `},
		{Code: `
function test(str: string) {
  str.match('[a-z');
}
    `},
		{Code: `
const text = 'something';
declare const search: RegExp;
text.match(search);
      `},
		{Code: `
const text = 'something';
declare const obj: { search: RegExp };
text.match(obj.search);
      `},
		{Code: `
const text = 'something';
declare function returnsRegexp(): RegExp;
text.match(returnsRegexp());
      `},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   `'something'.match(/thing/);`,
			Output: []string{`/thing/.exec('something');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code:   `'something'.match('^[a-z]+thing/?$');`,
			Output: []string{`/^[a-z]+thing\/?$/.exec('something');`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `
const text = 'something';
const search = /thing/;
text.match(search);
      `,
			Output: []string{`
const text = 'something';
const search = /thing/;
search.exec(text);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: `
const text = 'something';
const search = 'thing';
text.match(search);
      `,
			Output: []string{`
const text = 'something';
const search = 'thing';
RegExp(search).exec(text);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: `
function test(text: string, pattern: string) {
  text.match(pattern);
}
      `,
			Output: []string{`
function test(text: string, pattern: string) {
  RegExp(pattern).exec(text);
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      3,
					Column:    8,
				},
			},
		},
		{
			Code: `
function f(s: 'a' | 'b') {
  s.match('a');
}
      `,
			Output: []string{`
function f(s: 'a' | 'b') {
  /a/.exec(s);
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  s.match(/thing/);
}
      `,
			Output: []string{`
type SafeString = string & { __HTML_ESCAPED__: void };
function f(s: SafeString) {
  /thing/.exec(s);
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: `
function f<T extends 'a' | 'b'>(s: T) {
  s.match(/thing/);
}
      `,
			Output: []string{`
function f<T extends 'a' | 'b'>(s: T) {
  /thing/.exec(s);
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      3,
					Column:    5,
				},
			},
		},
		{
			Code: `
const text = 'something';
const search = new RegExp('test', '');
text.match(search);
      `,
			Output: []string{`
const text = 'something';
const search = new RegExp('test', '');
search.exec(text);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      4,
					Column:    6,
				},
			},
		},
		{
			Code: `
function test(pattern: string) {
  'check'.match(new RegExp(pattern, undefined));
}
      `,
			Output: []string{`
function test(pattern: string) {
  new RegExp(pattern, undefined).exec('check');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      3,
					Column:    11,
				},
			},
		},
		{
			Code:   "\nfunction temp(text: string): void {\n  text.match(new RegExp(`${'hello'}`));\n  text.match(new RegExp(`${'hello'.toString()}`));\n}\n      ",
			Output: []string{"\nfunction temp(text: string): void {\n  new RegExp(`${'hello'}`).exec(text);\n  new RegExp(`${'hello'.toString()}`).exec(text);\n}\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      3,
					Column:    8,
				},
				{
					MessageId: `regExpExecOverStringMatch`,
					Line:      4,
					Column:    8,
				},
			},
		},
	})
}
