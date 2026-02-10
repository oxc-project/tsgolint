package prefer_string_starts_ends_with

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestPreferStringStartsEndsWithRule(t *testing.T) {
	t.Parallel()

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferStringStartsEndsWithRule, []rule_tester.ValidTestCase{
		{
			Code: `
      function f(s: string[]) {
        s[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[] | null) {
        s?.[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[] | undefined) {
        s?.[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s[0] + 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s[1] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string | undefined) {
        s?.[1] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string | string[]) {
        s[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: any) {
        s[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f<T>(s: T) {
        s[0] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s[s.length - 1] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[] | undefined) {
        s?.[s.length - 1] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s[s.length - 2] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string | undefined) {
        s?.[s.length - 2] === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.charAt(0) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[] | undefined) {
        s?.charAt(0) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.charAt(0) + 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.charAt(1) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string | undefined) {
        s?.charAt(1) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.charAt() === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.charAt(s.length - 1) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(a: string, b: string, c: string) {
        (a + b).charAt((a + c).length - 1) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(a: string, b: string, c: string) {
        (a + b).charAt(c.length - 1) === 'a';
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.indexOf(needle) === 0;
      }
    `,
		},
		{
			Code: `
      function f(s: string | string[]) {
        s.indexOf(needle) === 0;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.indexOf(needle) === s.length - needle.length;
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.lastIndexOf(needle) === s.length - needle.length;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.lastIndexOf(needle) === 0;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^foo/);
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/foo$/);
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^foo/) + 1;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/foo$/) + 1;
      }
    `,
		},
		{
			Code: `
      function f(s: { match(x: any): boolean }) {
        s.match(/^foo/) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: { match(x: any): boolean }) {
        s.match(/foo$/) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/foo/) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^foo$/) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^foo./) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^foo|bar/) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(new RegExp('')) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(pattern) !== null; // cannot check '^'/'$'
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(new RegExp('^/!{[', 'u')) !== null; // has syntax error
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match() !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(777) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^bar/) !== undefined;
      }
    `,
		},
		{
			Code: `
      const missing = undefined;
      function f(s: string) {
        s.match(/^bar/) === missing;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.match(/^bar/g) !== null;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        /^bar/g.test(s);
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        /bar$/y.test(s);
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.slice(0, needle.length) === needle;
      }
    `,
		},
		{
			Code: `
      function f(s: string[]) {
        s.slice(-needle.length) === needle;
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(1, 4) === 'bar';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(-4, -1) === 'bar';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(1) === 'bar';
      }
    `,
		},
		{
			Code: `
      function f(s: string | null) {
        s?.slice(1) === 'bar';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        pattern.test(s);
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        /^bar/.test();
      }
    `,
		},
		{
			Code: `
      function f(x: { test(): void }, s: string) {
        x.test(s);
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(0, -4) === 'car';
      }
    `,
		},
		{
			Code: `
      function f(x: string, s: string) {
        x.endsWith('foo') && x.slice(0, -4) === 'bar';
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(0, length) === needle; // the 'length' can be different to 'needle.length'
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(-length) === needle; // 'length' can be different
      }
    `,
		},
		{
			Code: `
      function f(s: string) {
        s.slice(0, 3) === needle;
      }
    `,
		},
		{
			Code: `
        declare const s: string;
        s[0] === 'a';
      `,
			Options: rule_tester.OptionsFromJSON[PreferStringStartsEndsWithOptions](`{"allowSingleElementEquality":"always"}`),
		},
		{
			Code: `
        declare const s: string;
        s[s.length - 1] === 'a';
      `,
			Options: rule_tester.OptionsFromJSON[PreferStringStartsEndsWithOptions](`{"allowSingleElementEquality":"always"}`),
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
        function f(s: string) {
          s[0] === 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s?.[0] === 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s?.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[0] !== 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s?.[0] !== 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s?.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[0] == 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[0] != 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[0] === 'あ';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('あ');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[0] === '👍'; // the length is 2.
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string, t: string) {
          s[0] === t; // the length of t is unknown.
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s[s.length - 1] === 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          (s)[0] === ("a")
        }
      `,
			Output: []string{`
        function f(s: string) {
          (s).startsWith("a")
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) === 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) !== 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) == 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) != 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) === 'あ';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('あ');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(0) === '👍'; // the length is 2.
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string, t: string) {
          s.charAt(0) === t; // the length of t is unknown.
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.charAt(s.length - 1) === 'a';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('a');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          (s).charAt(0) === "a";
        }
      `,
			Output: []string{`
        function f(s: string) {
          (s).startsWith("a");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.indexOf(needle) === 0;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s?.indexOf(needle) === 0;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s?.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.indexOf(needle) !== 0;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.indexOf(needle) == 0;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.indexOf(needle) != 0;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf('bar') === s.length - 3;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf('bar') !== s.length - 3;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf('bar') == s.length - 3;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf('bar') != s.length - 3;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf('bar') === s.length - 'bar'.length;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.lastIndexOf(needle) === s.length - needle.length;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) !== null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s?.match(/^bar/) !== null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s?.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) != null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) != undefined;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/bar$/) !== null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/bar$/) != null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) === null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) == null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/^bar/) == undefined;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/bar$/) === null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.endsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.match(/bar$/) == null;
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.endsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = /^bar/;
        function f(s: string) {
          s.match(pattern) != null;
        }
      `,
			Output: []string{`
        const pattern = /^bar/;
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = new RegExp('^bar');
        function f(s: string) {
          s.match(pattern) != null;
        }
      `,
			Output: []string{`
        const pattern = new RegExp('^bar');
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = /^"quoted"/;
        function f(s: string) {
          s.match(pattern) != null;
        }
      `,
			Output: []string{`
        const pattern = /^"quoted"/;
        function f(s: string) {
          s.startsWith("\"quoted\"");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, 3) === 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s?.slice(0, 3) === 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s?.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, 3) !== 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, 3) == 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, 3) != 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, needle.length) === needle;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(0, needle.length) == needle; // hating implicit type conversion
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(-3) === 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(-3) !== 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          !s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(-needle.length) === needle;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.slice(s.length - needle.length) === needle;
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.substring(0, 3) === 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.substring(-3) === 'bar'; // the code is probably mistake.
        }
      `,
			Output: []string{},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          s.substring(s.length - 3, s.length) === 'bar';
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith('bar');
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          /^bar/.test(s);
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          /^bar/?.test(s);
        }
      `,
			Output: []string{`
        function f(s: string) {
          s?.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          /bar$/.test(s);
        }
      `,
			Output: []string{`
        function f(s: string) {
          s.endsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferEndsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = /^bar/;
        function f(s: string) {
          pattern.test(s);
        }
      `,
			Output: []string{`
        const pattern = /^bar/;
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = new RegExp('^bar');
        function f(s: string) {
          pattern.test(s);
        }
      `,
			Output: []string{`
        const pattern = new RegExp('^bar');
        function f(s: string) {
          s.startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        const pattern = /^"quoted"/;
        function f(s: string) {
          pattern.test(s);
        }
      `,
			Output: []string{`
        const pattern = /^"quoted"/;
        function f(s: string) {
          s.startsWith("\"quoted\"");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: string) {
          /^bar/.test(a + b);
        }
      `,
			Output: []string{`
        function f(s: string) {
          (a + b).startsWith("bar");
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f(s: 'a' | 'b') {
          s.indexOf(needle) === 0;
        }
      `,
			Output: []string{`
        function f(s: 'a' | 'b') {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        function f<T extends 'a' | 'b'>(s: T) {
          s.indexOf(needle) === 0;
        }
      `,
			Output: []string{`
        function f<T extends 'a' | 'b'>(s: T) {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
		{
			Code: `
        type SafeString = string & { __HTML_ESCAPED__: void };
        function f(s: SafeString) {
          s.indexOf(needle) === 0;
        }
      `,
			Output: []string{`
        type SafeString = string & { __HTML_ESCAPED__: void };
        function f(s: SafeString) {
          s.startsWith(needle);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: `preferStartsWith`,
				},
			},
		},
	})
}
