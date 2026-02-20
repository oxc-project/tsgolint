package dot_notation

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestDotNotation(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &DotNotationRule, []rule_tester.ValidTestCase{
		{Code: "a.b;"},
		{Code: "a.b.c;"},
		{Code: "a['12'];"},
		{Code: "a[b];"},
		{Code: "a[0];"},
		{Code: "a.b.c;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a.arguments;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a.let;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a.yield;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a.eval;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a[0];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a['while'];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a['true'];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a['null'];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a[true];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a[null];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a.true;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":true}")},
		{Code: "a.null;", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":true}")},
		{Code: "a['snake_case'];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}")},
		{Code: "a['lots_of_snake_case'];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}")},
		{Code: "a[`time${range}`];"},
		{Code: "a[`while`];", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}")},
		{Code: "a[`time range`];"},
		{Code: "a.true;"},
		{Code: "a.null;"},
		{Code: "a[undefined];"},
		{Code: "a[void 0];"},
		{Code: "a[b()];"},
		{Code: "a[/(?<zero>0)/];"},
		{Code: "\nclass X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx['priv_prop'] = 123;\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPrivateClassPropertyAccess\":true}")},
		{Code: "\nclass X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx['protected_prop'] = 123;\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowProtectedClassPropertyAccess\":true}")},
		{Code: "\nclass X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx['hello'] = 3;\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowIndexSignaturePropertyAccess\":true}")},
		{Code: "\ninterface Nested {\n  property: string;\n  [key: string]: number | string;\n}\n\nclass Dingus {\n  nested: Nested;\n}\n\nlet dingus: Dingus | undefined;\n\ndingus?.nested.property;\ndingus?.nested['hello'];\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowIndexSignaturePropertyAccess\":true}")},
		{Code: "\nclass X {\n  private priv_prop = 123;\n}\n\nlet x: X | undefined;\nconsole.log(x?.['priv_prop']);\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPrivateClassPropertyAccess\":true}")},
		{Code: "\nclass X {\n  protected priv_prop = 123;\n}\n\nlet x: X | undefined;\nconsole.log(x?.['priv_prop']);\n      ", Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowProtectedClassPropertyAccess\":true}")},
		{Code: "\ntype Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\ndeclare const foo: Foo;\nfoo['key_baz'];\n      ", TSConfig: "tsconfig.noPropertyAccessFromIndexSignature.json"},
		{Code: "\ntype Key = Lowercase<string>;\ntype Foo = {\n  BAR: boolean;\n  [key: Lowercase<string>]: number;\n};\ndeclare const foo: Foo;\nfoo['bar'];\n      ", TSConfig: "tsconfig.noPropertyAccessFromIndexSignature.json"},
		{Code: "\ntype ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x['extraKey'];\n}\n      ", TSConfig: "tsconfig.noPropertyAccessFromIndexSignature.json"},
		{Code: "\ntype Foo = {\n  [key: string]: number;\n};\ndeclare const foo: Foo;\nfoo[`key_baz`];\n      ", TSConfig: "tsconfig.noPropertyAccessFromIndexSignature.json"},
		// Hyphenated property names are not valid JS identifiers and must use bracket notation
		{Code: "a['X-Amzn-Trace-Id'];"},
		{Code: "a['X-Amzn-Trace-Id'];", Tsx: true},
		{Code: "a['has space'];"},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "\nclass X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx['priv_prop'] = 123;\n      ",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPrivateClassPropertyAccess\":false}"),
			Output:  []string{"\nclass X {\n  private priv_prop = 123;\n}\n\nconst x = new X();\nx.priv_prop = 123;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "\nclass X {\n  public pub_prop = 123;\n}\n\nconst x = new X();\nx['pub_prop'] = 123;\n      ",
			Output: []string{"\nclass X {\n  public pub_prop = 123;\n}\n\nconst x = new X();\nx.pub_prop = 123;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a['true'];",
			Output: []string{"a.true;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a['time'];",
			Output: []string{"a.time;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a[null];",
			Output: []string{"a.null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a[true];",
			Output: []string{"a.true;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a[false];",
			Output: []string{"a.false;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a['b'];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a['π'];",
			Output: []string{"a.π;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "a.b['c'];",
			Output: []string{"a.b.c;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:    "a['_dangle'];",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}"),
			Output:  []string{"a._dangle;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:    "a['SHOUT_CASE'];",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowPattern\":\"^[a-z]+(_[a-z]+)+$\"}"),
			Output:  []string{"a.SHOUT_CASE;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "\na\n  ['SHOUT_CASE'];\n      ",
			Output: []string{"\na\n  .SHOUT_CASE;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot", Line: 3, Column: 4},
			},
		},
		{
			Code:   "getResource()\n    .then(function(){})\n    [\"catch\"](function(){})\n    .then(function(){})\n    [\"catch\"](function(){});",
			Output: []string{"getResource()\n    .then(function(){})\n    .catch(function(){})\n    .then(function(){})\n    .catch(function(){});"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot", Line: 3, Column: 6},
				{MessageId: "useDot", Line: 5, Column: 6},
			},
		},
		{
			Code:    "\nfoo\n  .while;\n      ",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}"),
			Output:  []string{"\nfoo\n  [\"while\"];\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useBrackets"},
			},
		},
		{
			Code:   "foo[/* comment */ 'bar'];",
			Output: []string{"foo.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "foo['bar' /* comment */];",
			Output: []string{"foo.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "foo['bar'];",
			Output: []string{"foo.bar;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:    "foo./* comment */ while;",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}"),
			Output:  []string{"foo/* comment */ [\"while\"];"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useBrackets"},
			},
		},
		{
			Code:   "foo[null];",
			Output: []string{"foo.null;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "foo['bar'] instanceof baz;",
			Output: []string{"foo.bar instanceof baz;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:    "let.if();",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowKeywords\":false}"),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useBrackets"},
			},
		},
		{
			Code:    "\nclass X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx['protected_prop'] = 123;\n      ",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowProtectedClassPropertyAccess\":false}"),
			Output:  []string{"\nclass X {\n  protected protected_prop = 123;\n}\n\nconst x = new X();\nx.protected_prop = 123;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:    "\nclass X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx['prop'] = 'hello';\n      ",
			Options: rule_tester.OptionsFromJSON[DotNotationOptions]("{\"allowIndexSignaturePropertyAccess\":true}"),
			Output:  []string{"\nclass X {\n  prop: string;\n  [key: string]: number;\n}\n\nconst x = new X();\nx.prop = 'hello';\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "\ntype Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\nfoo['key_baz'];\n      ",
			Output: []string{"\ntype Foo = {\n  bar: boolean;\n  [key: `key_${string}`]: number;\n};\nfoo.key_baz;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "\ntype ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x['extraKey'];\n}\n      ",
			Output: []string{"\ntype ExtraKey = `extra${string}`;\n\ntype Foo = {\n  foo: string;\n  [extraKey: ExtraKey]: number;\n};\n\nfunction f<T extends Foo>(x: T) {\n  x.extraKey;\n}\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
		},
		{
			Code:   "\ntype Foo = {\n  prop: boolean;\n  [key: string]: number | boolean;\n};\ndeclare const foo: Foo;\nfoo[`prop`];\n      ",
			Output: []string{"\ntype Foo = {\n  prop: boolean;\n  [key: string]: number | boolean;\n};\ndeclare const foo: Foo;\nfoo.prop;\n      "},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
			},
			TSConfig: "tsconfig.noPropertyAccessFromIndexSignature.json",
		},
	})
}
