package no_unsafe_assignment

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/rule"
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestDiagnosticRelationships(t *testing.T) {
	t.Parallel()

	primaryRange := core.NewTextRange(10, 11)
	senderRange := core.NewTextRange(20, 30)
	receiverRange := core.NewTextRange(1, 8)
	message := rule.RuleMessage{Id: "test", Description: "test"}

	assignment := buildAssignmentDiagnostic(primaryRange, senderRange, receiverRange, "Set<any>", "Set<string>", message)
	if assignment.Range != primaryRange {
		t.Fatalf("assignment range = %v, want %v", assignment.Range, primaryRange)
	}
	if len(assignment.LabeledRanges) != 2 {
		t.Fatalf("assignment labels = %d, want 2", len(assignment.LabeledRanges))
	}
	if assignment.LabeledRanges[0].Range != senderRange || assignment.LabeledRanges[0].Label != "Assigned value has type `Set<any>`." {
		t.Fatalf("sender label = %+v", assignment.LabeledRanges[0])
	}
	if assignment.LabeledRanges[1].Range != receiverRange || assignment.LabeledRanges[1].Label != "Target expects type `Set<string>`." {
		t.Fatalf("receiver label = %+v", assignment.LabeledRanges[1])
	}

	thisAssignment := buildThisAssignmentDiagnostic(primaryRange, senderRange, receiverRange, "any", "number", message)
	if thisAssignment.LabeledRanges[0].Label != "`this` has type `any`." || thisAssignment.LabeledRanges[0].Range != senderRange {
		t.Fatalf("this label = %+v", thisAssignment.LabeledRanges[0])
	}

	destructure := buildDestructureDiagnostic(receiverRange, senderRange, "[any]", "any", message)
	if destructure.Range != receiverRange || len(destructure.LabeledRanges) != 2 {
		t.Fatalf("destructure diagnostic = %+v", destructure)
	}
	if destructure.LabeledRanges[0].Label != "Destructured source provides type `[any]`." || destructure.LabeledRanges[0].Range != senderRange {
		t.Fatalf("destructure source label = %+v", destructure.LabeledRanges[0])
	}
	if destructure.LabeledRanges[1].Label != "This binding receives type `any`." || destructure.LabeledRanges[1].Range != receiverRange {
		t.Fatalf("destructure binding label = %+v", destructure.LabeledRanges[1])
	}

	spread := buildArraySpreadDiagnostic(primaryRange, senderRange, "any[]", message)
	if spread.Range != primaryRange || len(spread.LabeledRanges) != 1 || spread.LabeledRanges[0].Range != senderRange || spread.LabeledRanges[0].Label != "Spread value has type `any[]`." {
		t.Fatalf("spread diagnostic = %+v", spread)
	}
}

func assignmentTest(
	tests []struct {
		code               string
		col                int
		endCol             int
		skipAssignmentExpr bool
	},
) []rule_tester.InvalidTestCase {
	res := make([]rule_tester.InvalidTestCase, 0, 3*len(tests))
	for _, test := range tests {
		res = append(res,
			// VariableDeclaration
			rule_tester.InvalidTestCase{
				Code: "const " + test.code,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 6,
						EndColumn: test.endCol + 6,
						Line:      1,
						MessageId: "unsafeArrayPatternFromTuple",
					},
				},
			},
			// AssignmentPattern
			rule_tester.InvalidTestCase{
				Code: "function foo(" + test.code + ") {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 13,
						EndColumn: test.endCol + 13,
						Line:      1,
						MessageId: "unsafeArrayPatternFromTuple",
					},
				},
			},
		)
		if !test.skipAssignmentExpr {
			// AssignmentExpression
			res = append(res, rule_tester.InvalidTestCase{
				Code: "(" + test.code + ")",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						Column:    test.col + 1,
						EndColumn: test.endCol + 1,
						Line:      1,
						MessageId: "unsafeArrayPatternFromTuple",
					},
				},
			})
		}
	}
	return res
}

func TestNoUnsafeAssignmentRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.noImplicitThis.json", t, &NoUnsafeAssignmentRule, []rule_tester.ValidTestCase{
		{Code: "const x = 1;"},
		{Code: "const x: number = 1;"},
		{Code: `
const x = 1,
  y = 1;
    `},
		{Code: "let x;"},
		{Code: `
let x = 1,
  y;
    `},
		{Code: "function foo(a = 1) {}"},
		{Code: `
class Foo {
  constructor(private a = 1) {}
}
    `},
		{Code: `
class Foo {
  private a = 1;
}
    `},
		{Code: `
class Foo {
  accessor a = 1;
}
    `},
		{Code: "const x: Set<string> = new Set();"},
		{Code: "const x: Set<string> = new Set<string>();"},
		{Code: "const [x] = [1];"},
		{Code: "const [x, y] = [1, 2] as number[];"},
		{Code: "const [x, ...y] = [1, 2, 3, 4, 5];"},
		{Code: "const [x, ...y] = [1];"},
		{Code: "const [{ ...x }] = [{ x: 1 }] as [{ x: any }];"},
		{Code: "function foo(x = 1) {}"},
		{Code: "function foo([x] = [1]) {}"},
		{Code: "function foo([x, ...y] = [1, 2, 3, 4, 5]) {}"},
		{Code: "function foo([x, ...y] = [1]) {}"},
		{Code: "const x = new Set<any>();"},
		{Code: "const x = { y: 1 };"},
		// TODO(port): this is invalid TypeScript code
		{Skip: true, Code: "const x = { y = 1 };"},
		{Code: "const x = { y(){} };"},
		{Code: "const x: { y: number } = { y: 1 };"},
		{Code: "const x = [...[1, 2, 3]];"},
		{Code: "const [{ [`x${1}`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];"},
		{Code: `
type T = [string, T[]];
const test: T = ['string', []] as T;
    `},
		{
			Code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={'foo'} />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { a: string }): never;
<Foo a="foo" />;
      `,
			Tsx: true,
		},
		{
			Code: `
declare function Foo(props: { a: string }): never;
<Foo a={} />;
      `,
			Tsx: true,
		},
		{Code: "const x: unknown = y as any;"},
		{Code: "const x: unknown[] = y as any[];"},
		{Code: "const x: Set<unknown> = y as Set<any>;"},
		{Code: "const x: Map<string, string> = new Map();"},
		{Code: `
type Foo = { bar: unknown };
const bar: any = 1;
const foo: Foo = { bar };
    `},

		{Code: `
declare const foo: any;
let a = 1;

a+= foo;
		`},
	}, slices.Concat([]rule_tester.InvalidTestCase{
		{
			Code: "const x = 1 as any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
					Column:    9,
					EndColumn: 10,
				},
			},
		},
		{
			Code: `
const x = 1 as any,
  y = 1;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: "function foo(a = 1 as any) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  constructor(private a = 1 as any) {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  private a = 1 as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
class Foo {
  accessor a = 1 as any;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [x] = spooky;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [[[x]]] = [spooky];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArrayPatternFromTuple",
				},
			},
		},
		{
			Code: `
const {
  x: { y: z },
} = { x: spooky };
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArrayPatternFromTuple",
				},
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
let value: number;

value = spooky;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
					Line:      4,
					Column:    7,
					EndColumn: 8,
				},
			},
		},
		{
			Code: `
const [x] = 1 as any;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyAssignment",
				},
			},
		},
		{
			Code: `
const [x] = [] as any[];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeArrayPattern",
				},
			},
		},
		{
			Code: "const x: Set<string> = new Set<any>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
					Column:    22,
					EndColumn: 23,
				},
			},
		},
		{
			Code: "const x: Map<string, string> = new Map<string, any>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
		{
			Code: "const x: Set<string[]> = new Set<any[]>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
		{
			Code: "const x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>();",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unsafeAssignment",
				},
			},
		},
	},
		assignmentTest([]struct {
			code               string
			col                int
			endCol             int
			skipAssignmentExpr bool
		}{
			{"[x] = [1] as [any]", 2, 3, false},
			{"[[[[x]]]] = [[[[1 as any]]]]", 5, 6, false},
			{"[[[[x]]]] = [1 as any]", 2, 9, true},
			{"[{x}] = [{x: 1}] as [{x: any}]", 3, 4, false},
			{"[{['x']: x}] = [{['x']: 1}] as [{['x']: any}]", 10, 11, false},
			{"[{[`x`]: x}] = [{[`x`]: 1}] as [{[`x`]: any}]", 10, 11, false},
		}),

		[]rule_tester.InvalidTestCase{
			{
				Code: "[[[[x]]]] = [1 as any];",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeAssignment",
						Line:      1,
						Column:    11,
						EndColumn: 12,
					},
				},
			},
			{
				Code: `
const x = [...(1 as any)];
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArraySpread",
						Line:      2,
						Column:    12,
						EndColumn: 15,
					},
				},
			},
			{
				Code: `
const x = [...([] as any[])];
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeArraySpread",
						Line:      2,
						Column:    12,
						EndColumn: 15,
					},
				},
			},
		},
		assignmentTest([]struct {
			code               string
			col                int
			endCol             int
			skipAssignmentExpr bool
		}{
			{"{x} = {x: 1} as {x: any}", 2, 3, false},
			{"{x: y} = {x: 1} as {x: any}", 5, 6, false},
			{"{x: {y}} = {x: {y: 1}} as {x: {y: any}}", 6, 7, false},
			{"{x: [y]} = {x: {y: 1}} as {x: [any]}", 6, 7, false},
		}),

		[]rule_tester.InvalidTestCase{
			{
				Code: "const x = { y: 1 as any };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    14,
						EndColumn: 15,
					},
				},
			},
			{
				Code: "const x = { y: { z: 1 as any } };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    19,
						EndColumn: 20,
					},
				},
			},
			{
				Code: "const x: { y: Set<Set<Set<string>>> } = { y: new Set<Set<Set<any>>>() };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unsafeAssignment",
						Column:    44,
						EndColumn: 45,
					},
				},
			},
			{
				Code: "const x = { ...(1 as any) };",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Column:    9,
						EndColumn: 10,
					},
				},
			},
			{
				Code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={1 as any} />;
      `,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    7,
						EndColumn: 8,
					},
				},
			},
			{
				Code: `
function foo() {
  const bar = this;
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignmentThis",
						Line:      3,
						Column:    13,
						EndColumn: 14,
					},
				},
			},
			{
				Code: `
type T = [string, T[]];
const test: T = ['string', []] as any;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      3,
						Column:    15,
						EndColumn: 16,
					},
				},
			},
			{
				Code: `
type Foo = { bar: number };
const bar: any = 1;
const foo: Foo = { bar };
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    20,
						EndColumn: 23,
					},
				},
			},

			{
				Code: `
declare const foo: any;
interface Bar {
  bar: number
}

class Foo {
  constructor(
    private readonly param: Bar = Object.create(null)
  ) {}
}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      9,
						Column:    33,
						EndColumn: 34,
					},
				},
			},
			{
				Code: `
let foo: { foo: 1 };

foo = { bar: 2 } as any;
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "anyAssignment",
						Line:      4,
						Column:    5,
						EndColumn: 6,
					},
				},
			},
		}))
}
