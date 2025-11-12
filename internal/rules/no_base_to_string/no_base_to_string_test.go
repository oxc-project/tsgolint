package no_base_to_string

import (
	"fmt"
	"slices"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func FlatMap[A, B any](input []A, f func(A) []B) []B {
	var result []B
	for _, v := range input {
		result = append(result, f(v)...)
	}
	return result
}
func TestNoBaseToStringRule(t *testing.T) {
	literalListBasic := []string{
		"''",
		"'text'",
		"true",
		"false",
		"1",
		"1n",
		"[]",
		"/regex/",
	}

	literalListNeedParen := []string{
		"__dirname === 'foobar'",
		"{}.constructor()",
		"() => {}",
		"function() {}",
	}

	literalList := slices.Concat(literalListBasic, literalListNeedParen)

	literalListWrapped := slices.Concat(
		literalListBasic,
		utils.Map(literalListNeedParen, func(l string) string { return fmt.Sprintf("(%v)", l) }),
	)

	extraValid := utils.Map(slices.Concat(
		// template
		utils.Map(literalList, func(i string) string {
			return fmt.Sprintf("`${%v}`;", i)
		}),

		// operator + +=
		FlatMap(literalListWrapped, func(l string) []string {
			return utils.Map(literalListWrapped, func(r string) string {
				return fmt.Sprintf("%v + %v;", l, r)
			})
		}),

		// toString()
		utils.Map(literalListWrapped, func(i string) string {
			if i == "1" {
				i = "(1)"
			}
			return fmt.Sprintf("%v.toString();", i)
		}),

		// variable toString() and template
		utils.Map(literalList, func(i string) string {
			return `
        let value = ` + i + `;
        value.toString();
        let text = ` + "`${value}`;\n"
		}),

		// String()
		utils.Map(literalList, func(i string) string { return "String(" + i + ");" }),
	), func(s string) rule_tester.ValidTestCase {
		return rule_tester.ValidTestCase{
			Code: s,
		}
	})

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoBaseToStringRule, slices.Concat(
		extraValid,
		[]rule_tester.ValidTestCase{

			{Code: `
function someFunction() {}
someFunction.toString();
let text = ` + "`" + `${someFunction}` + "`" + `;
    `},
			{Code: `
function someFunction() {}
someFunction.toLocaleString();
let text = ` + "`" + `${someFunction}` + "`" + `;
    `},
			{Code: "unknownObject.toString();"},
			{Code: "unknownObject.toLocaleString();"},
			{Code: "unknownObject.someOtherMethod();"},
			{Code: `
class CustomToString {
  toString() {
    return 'Hello, world!';
  }
}
'' + new CustomToString();
    `},
			{Code: `
const literalWithToString = {
  toString: () => 'Hello, world!',
};
'' + literalWithToString;
    `},
			{Code: `
const printer = (inVar: string | number | boolean) => {
  inVar.toString();
};
printer('');
printer(1);
printer(true);
    `},
			{Code: `
const printer = (inVar: string | number | boolean) => {
  inVar.toLocaleString();
};
printer('');
printer(1);
printer(true);
    `},
			{Code: "let _ = {} * {};"},
			{Code: "let _ = {} / {};"},
			{Code: "let _ = ({} *= {});"},
			{Code: "let _ = ({} /= {});"},
			{Code: "let _ = ({} = {});"},
			{Code: "let _ = {} == {};"},
			{Code: "let _ = {} === {};"},
			{Code: "let _ = {} in {};"},
			{Code: "let _ = {} & {};"},
			{Code: "let _ = {} ^ {};"},
			{Code: "let _ = {} << {};"},
			{Code: "let _ = {} >> {};"},
			{Code: `
function tag() {}
tag` + "`" + `${{}}` + "`" + `;
    `},
			{Code: `
      function tag() {}
      tag` + "`" + `${{}}` + "`" + `;
    `},
			{Code: `
      interface Brand {}
      function test(v: string & Brand): string {
        return ` + "`" + `${v}` + "`" + `;
      }
    `},
			{Code: "'' += new Error();"},
			{Code: "'' += new URL();"},
			{Code: "'' += new URLSearchParams();"},
			{Code: `
Number(1);
    `},
			{
				Code:    "String(/regex/);",
				Options: NoBaseToStringOptions{IgnoredTypeNames: []string{"RegExp"}},
			},
			{
				Code: `
type Foo = { a: string } | { b: string };
declare const foo: Foo;
String(foo);
      `,
				Options: NoBaseToStringOptions{IgnoredTypeNames: []string{"Foo"}},
			},
			// TODO(port): this is invalid ts file (with lib)
			{Code: `
function String(value) {
  return value;
}
declare const myValue: object;
String(myValue);
`, Skip: true},
			{Code: `
import { String } from 'foo';
String({});
    `},
			{Code: `
['foo', 'bar'].join('');
    `},
			{Code: `
([{ foo: 'foo' }, 'bar'] as string[]).join('');
    `},
			{Code: `
function foo<T extends string>(array: T[]) {
  return array.join();
}
    `},
			{Code: `
class Foo {
  toString() {
    return '';
  }
}
[new Foo()].join();
    `},
			{Code: `
class Foo {
  join() {}
}
const foo = new Foo();
foo.join();
    `},
			{Code: `
declare const array: string[];
array.join('');
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
array.join('');
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
array.join('');
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
array.join('');
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
tuple.join('');
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
tuple.join('');
    `},
			{Code: `
String(['foo', 'bar']);
    `},
			{Code: `
String([{ foo: 'foo' }, 'bar'] as string[]);
    `},
			{Code: `
function foo<T extends string>(array: T[]) {
  return String(array);
}
    `},
			{Code: `
class Foo {
  toString() {
    return '';
  }
}
String([new Foo()]);
    `},
			{Code: `
declare const array: string[];
String(array);
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
String(array);
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
String(array);
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
String(array);
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
String(tuple);
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
String(tuple);
    `},
			{Code: `
['foo', 'bar'].toString();
    `},
			{Code: `
([{ foo: 'foo' }, 'bar'] as string[]).toString();
    `},
			{Code: `
function foo<T extends string>(array: T[]) {
  return array.toString();
}
    `},
			{Code: `
class Foo {
  toString() {
    return '';
  }
}
[new Foo()].toString();
    `},
			{Code: `
declare const array: string[];
array.toString();
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
array.toString();
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
array.toString();
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
array.toString();
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
tuple.toString();
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
tuple.toString();
    `},
			{Code: `
` + "`" + `${['foo', 'bar']}` + "`" + `;
    `},
			{Code: `
` + "`" + `${[{ foo: 'foo' }, 'bar'] as string[]}` + "`" + `;
    `},
			{Code: `
function foo<T extends string>(array: T[]) {
  return ` + "`" + `${array}` + "`" + `;
}
    `},
			{Code: `
class Foo {
  toString() {
    return '';
  }
}
` + "`" + `${[new Foo()]}` + "`" + `;
    `},
			{Code: `
declare const array: string[];
` + "`" + `${array}` + "`" + `;
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const array: (string & Foo)[];
` + "`" + `${array}` + "`" + `;
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] | (string & Bar)[];
` + "`" + `${array}` + "`" + `;
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const array: (string & Foo)[] & (string & Bar)[];
` + "`" + `${array}` + "`" + `;
    `},
			{Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const tuple: [string & Foo, string & Bar];
` + "`" + `${tuple}` + "`" + `;
    `},
			{Code: `
class Foo {
  foo: string;
}
declare const tuple: [string] & [Foo];
` + "`" + `${tuple}` + "`" + `;
    `},
			{Code: `
let objects = [{}, {}];
String(...objects);
    `},
			{Code: `
type Constructable<Entity> = abstract new (...args: any[]) => Entity;

interface GuildChannel {
  toString(): ` + "`" + `<#${string}>` + "`" + `;
}

declare const foo: Constructable<GuildChannel & { bar: 1 }>;
class ExtendedGuildChannel extends foo {}
declare const bb: ExtendedGuildChannel;
bb.toString();
    `},
			{Code: `
type Constructable<Entity> = abstract new (...args: any[]) => Entity;

interface GuildChannel {
  toString(): ` + "`" + `<#${string}>` + "`" + `;
}

declare const foo: Constructable<{ bar: 1 } & GuildChannel>;
class ExtendedGuildChannel extends foo {}
declare const bb: ExtendedGuildChannel;
bb.toString();
    `},
			{Code: `
function foo<T>(x: T) {
  String(x);
}
    `},
			{Code: `
declare const u: unknown;
String(u);
    `},
			{Code: `
type Value = string | Value[];
declare const v: Value;

String(v);
    `},
			{Code: `
type Value = (string | Value)[];
declare const v: Value;

String(v);
    `},
			{Code: `
type Value = Value[];
declare const v: Value;

String(v);
    `},
			{Code: `
type Value = [Value];
declare const v: Value;

String(v);
    `},
			{Code: `
declare const v: ('foo' | 'bar')[][];
String(v);
    `},
		}), []rule_tester.InvalidTestCase{
		{
			Code: "`${{}})`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: "({}).toString();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: "({}).toLocaleString();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: "'' + {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: "String({});",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: "'' += {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString.toLocaleString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrString = Math.random() ? { a: true } : 'text';
        someObjectOrString + '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject.toLocaleString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        let someObjectOrObject = Math.random() ? { a: true, b: true } : { a: true };
        someObjectOrObject + '';
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        interface A {}
        interface B {}
        function test(intersection: A & B): string {
          return ` + "`" + `${intersection}` + "`" + `;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
class Foo {
  foo: string;
}
declare const foo: string | Foo;
` + "`" + `${foo}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const foo: Bar | Foo;
` + "`" + `${foo}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
class Foo {
  foo: string;
}
class Bar {
  bar: string;
}
declare const foo: Bar & Foo;
` + "`" + `${foo}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        [{}, {}].join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        const array = [{}, {}];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class A {
          a: string;
        }
        [new A(), 'str'].join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        tuple.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        tuple.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        tuple.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        tuple.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        tuple.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        const array = ['string', { foo: 'bar' }];
        array.join('');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array.join();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        String([{}, {}]);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = [{}, {}];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class A {
          a: string;
        }
        String([new A(), 'str']);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        String(tuple);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        String(tuple);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        String(tuple);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        String(tuple);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        String(tuple);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = ['string', { foo: 'bar' }];
        String(array);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return String(array);
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        [{}, {}].toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = [{}, {}];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class A {
          a: string;
        }
        [new A(), 'str'].toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        tuple.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        tuple.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        tuple.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        tuple.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        tuple.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = ['string', { foo: 'bar' }];
        array.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array.toString();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        ` + "`" + `${[{}, {}]}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = [{}, {}];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class A {
          a: string;
        }
        ` + "`" + `${[new A(), 'str']}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string | Foo)[];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: (string & Foo) | (string | Foo)[];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        class Bar {
          bar: string;
        }
        declare const array: Foo[] & Bar[];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const array: string[] | Foo[];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, Foo];
        ` + "`" + `${tuple}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, Foo];
        ` + "`" + `${tuple}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo | string, string];
        ` + "`" + `${tuple}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [string, string] | [Foo, Foo];
        ` + "`" + `${tuple}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        class Foo {
          foo: string;
        }
        declare const tuple: [Foo, string] & [Foo, Foo];
        ` + "`" + `${tuple}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        const array = ['string', { foo: 'bar' }];
        ` + "`" + `${array}` + "`" + `;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return ` + "`" + `${array}` + "`" + `;
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          array[0].toString();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(value: T) {
          value.toString();
        }
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
type Bar = Record<string, string>;
declare const foo: Bar | string;
foo.toString();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array;
        }
        foo([{ foo: 'foo' }]).join();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
        type Bar = Record<string, string>;
        function foo<T extends string | Bar>(array: T[]) {
          return array;
        }
        foo([{ foo: 'foo' }, 'bar']).join();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
		{
			Code: `
type Value = { foo: string } | Value[];
declare const v: Value;

String(v);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
type Value = ({ foo: string } | Value)[];
declare const v: Value;

String(v);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
type Value = [{ foo: string }, Value];
declare const v: Value;

String(v);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseToString",
				},
			},
		},
		{
			Code: `
declare const v: { foo: string }[][];
v.join();
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "baseArrayJoin",
				},
			},
		},
	})
}
