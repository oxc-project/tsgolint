package no_explicit_any

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExplicitAnyRule, []rule_tester.ValidTestCase{
		{
			Code: `
// Valid cases - no explicit any
const value: string = "hello";
let data: number = 42;
var info: object = {};

function processData(data: string) {
  return data;
}

function getData(): string {
  return "hello";
}

function processArgs(...args: string[]) {
  return args;
}

class Example {
  method(param: string): string {
    return param;
  }
  
  getData(): string {
    return "data";
  }
}

interface Config {
  data: string;
  options: number;
}

type DataType = string;
type ConfigType = number;

const typedValue: string = "value";
const array: string[] = [];
const object: { [key: string]: string } = {};

function genericFunction<T>(param: T): T {
  return param;
}

interface TestInterface {
  prop1: string;
  prop2: string[];
  prop3: { [key: string]: string };
}
`,
		},
		{
			Code: `
// Valid - using unknown instead of any
const value: unknown = "hello";
function processData(data: unknown) {
  return data;
}
function getData(): unknown {
  return "hello";
}
`,
		},
		{
			Code: `
// Valid - using proper types
interface User {
  name: string;
  age: number;
}

function processUser(user: User) {
  return user.name;
}

type UserData = User;
`,
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "const value: any = 'hello';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      1,
					Column:    14,
				},
			},
		},
		{
			Code: "let data: any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      1,
					Column:    12,
				},
			},
		},
		{
			Code: "var info: any = {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      1,
					Column:    13,
				},
			},
		},
		{
			Code: `
function processData(data: any) {
  return data;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      2,
					Column:    26,
				},
			},
		},
		{
			Code: `
function getData(): any {
  return "hello";
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      2,
					Column:    19,
				},
			},
		},
		{
			Code: `
function processArgs(...args: any[]) {
  return args;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAnyWithIgnoreRestArgs",
					Line:      2,
					Column:    26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "explicitAnyWithIgnoreRestArgsSuggestion",
							Output: `
function processArgs(...args: never[]) {
  return args;
}`,
						},
					},
				},
			},
		},
		{
			Code: `
class Example {
  method(param: any): any {
    return param;
  }
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      3,
					Column:    16,
				},
				{
					MessageId: "explicitAny",
					Line:      3,
					Column:    23,
				},
			},
		},
		{
			Code: `
interface Config {
  data: any;
  options: any;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      3,
					Column:  9,
				},
				{
					MessageId: "explicitAny",
					Line:      4,
					Column:  12,
				},
			},
		},
		{
			Code: `
type DataType = any;
type ConfigType = any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      2,
					Column:  16,
				},
				{
					MessageId: "explicitAny",
					Line:      3,
					Column:  18,
				},
			},
		},
		{
			Code: `
const typedValue: any = "value";
const array: any[] = [];
const object: { [key: string]: any } = {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      2,
					Column:    19,
				},
				{
					MessageId: "explicitAny",
					Line:      3,
					Column:    16,
				},
				{
					MessageId: "explicitAny",
					Line:      4,
					Column:    33,
				},
			},
		},
		{
			Code: `
function genericFunction<T = any>(param: T): T {
  return param;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "explicitAny",
					Line:      2,
					Column:    29,
				},
			},
		},
	})
}
