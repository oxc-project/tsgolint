package prefer_optional_chain

import (
	"fmt"
	"strings"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
)

// BaseCase represents a single test case from upstream typescript-eslint
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/prefer-optional-chain/base-cases.ts
type BaseCase struct {
	ID          int
	Chain       string // The input chain expression (with ${operator} placeholder)
	Declaration string // Type declarations
	OutputChain string // Expected output
}

// RawBaseCases returns all 26 base cases from upstream
// These are the core test cases that get transformed with different operators and mutations
func RawBaseCases() []BaseCase {
	return []BaseCase{
		// chained members
		{
			ID:          1,
			Chain:       "foo ${operator} foo.bar;",
			Declaration: "declare const foo: {bar: number} | null | undefined;",
			OutputChain: "foo?.bar;",
		},
		{
			ID:          2,
			Chain:       "foo.bar ${operator} foo.bar.baz;",
			Declaration: "declare const foo: {bar: {baz: number} | null | undefined};",
			OutputChain: "foo.bar?.baz;",
		},
		{
			ID:          3,
			Chain:       "foo ${operator} foo();",
			Declaration: "declare const foo: (() => number) | null | undefined;",
			OutputChain: "foo?.();",
		},
		{
			ID:          4,
			Chain:       "foo.bar ${operator} foo.bar();",
			Declaration: "declare const foo: {bar: (() => number) | null | undefined};",
			OutputChain: "foo.bar?.();",
		},
		{
			ID:          5,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.buzz;",
		},
		{
			ID:          6,
			Chain:       "foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined};",
			OutputChain: "foo.bar?.baz?.buzz;",
		},
		// case with a jump (i.e. a non-nullish prop)
		{
			ID:          7,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number}} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz.buzz;",
		},
		{
			ID:          8,
			Chain:       "foo.bar ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number}} | null | undefined};",
			OutputChain: "foo.bar?.baz.buzz;",
		},
		// case where for some reason there is a doubled up expression
		{
			ID:          9,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.buzz;",
		},
		{
			ID:          10,
			Chain:       "foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz;",
			Declaration: "declare const foo: {bar: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo.bar?.baz?.buzz;",
		},
		// chained members with element access
		{
			ID:          11,
			Chain:       "foo ${operator} foo[bar] ${operator} foo[bar].baz ${operator} foo[bar].baz.buzz;",
			Declaration: "declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.[bar]?.baz?.buzz;",
		},
		{
			ID:          12,
			Chain:       "foo ${operator} foo[bar].baz ${operator} foo[bar].baz.buzz;",
			Declaration: "declare const bar: string;\ndeclare const foo: {[k: string]: {baz: {buzz: number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.[bar].baz?.buzz;",
		},
		// case with a property access in computed property
		{
			ID:          13,
			Chain:       "foo ${operator} foo[bar.baz] ${operator} foo[bar.baz].buzz;",
			Declaration: "declare const bar: {baz: string};\ndeclare const foo: {[k: string]: {buzz: number} | null | undefined} | null | undefined;",
			OutputChain: "foo?.[bar.baz]?.buzz;",
		},
		// chained calls
		{
			ID:          14,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: () => number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.buzz();",
		},
		{
			ID:          15,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.buzz?.();",
		},
		{
			ID:          16,
			Chain:       "foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz.buzz ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined} | null | undefined} | null | undefined};",
			OutputChain: "foo.bar?.baz?.buzz?.();",
		},
		// case with a jump (i.e. a non-nullish prop)
		{
			ID:          17,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz.buzz();",
		},
		{
			ID:          18,
			Chain:       "foo.bar ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: () => number}} | null | undefined};",
			OutputChain: "foo.bar?.baz.buzz();",
		},
		{
			ID:          19,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz.buzz ${operator} foo.bar.baz.buzz();",
			Declaration: "declare const foo: {bar: {baz: {buzz: (() => number) | null | undefined}} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz.buzz?.();",
		},
		{
			ID:          20,
			Chain:       "foo.bar ${operator} foo.bar() ${operator} foo.bar().baz ${operator} foo.bar().baz.buzz ${operator} foo.bar().baz.buzz();",
			Declaration: "declare const foo: {bar: () => ({baz: {buzz: (() => number) | null | undefined} | null | undefined}) | null | undefined};",
			OutputChain: "foo.bar?.()?.baz?.buzz?.();",
		},
		// chained calls with element access
		{
			ID:          21,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz[buzz]();",
			Declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: () => number} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.[buzz]();",
		},
		{
			ID:          22,
			Chain:       "foo ${operator} foo.bar ${operator} foo.bar.baz ${operator} foo.bar.baz[buzz] ${operator} foo.bar.baz[buzz]();",
			Declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.[buzz]?.();",
		},
		// (partially) pre-optional chained
		{
			ID:          23,
			Chain:       "foo ${operator} foo?.bar ${operator} foo?.bar.baz ${operator} foo?.bar.baz[buzz] ${operator} foo?.bar.baz[buzz]();",
			Declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: (() => number) | null | undefined} | null | undefined} | null | undefined} | null | undefined;",
			OutputChain: "foo?.bar?.baz?.[buzz]?.();",
		},
		{
			ID:          24,
			Chain:       "foo ${operator} foo?.bar.baz ${operator} foo?.bar.baz[buzz];",
			Declaration: "declare const buzz: string;\ndeclare const foo: {bar: {baz: {[k: string]: number} | null | undefined}} | null | undefined;",
			OutputChain: "foo?.bar.baz?.[buzz];",
		},
		{
			ID:          25,
			Chain:       "foo ${operator} foo?.() ${operator} foo?.().bar;",
			Declaration: "declare const foo: (() => ({bar: number} | null | undefined)) | null | undefined;",
			OutputChain: "foo?.()?.bar;",
		},
		{
			ID:          26,
			Chain:       "foo.bar ${operator} foo.bar?.() ${operator} foo.bar?.().baz;",
			Declaration: "declare const foo: {bar: () => ({baz: number} | null | undefined)};",
			OutputChain: "foo.bar?.()?.baz;",
		},
	}
}

// MutateFn is a function that transforms a string
type MutateFn func(string) string

// Identity returns the input unchanged
func Identity(s string) string {
	return s
}

// BaseCaseOptions configures how base cases are generated
type BaseCaseOptions struct {
	Operator           string         // "&&" or "||"
	MutateCode         MutateFn       // Transform the input code
	MutateDeclaration  MutateFn       // Transform the declaration
	MutateOutput       MutateFn       // Transform the output (defaults to MutateCode)
	SkipIDs            map[int]bool   // IDs to skip
	UseSuggestionFixer bool           // Use suggestion instead of direct fix
	Options            map[string]any // Rule options
}

// GenerateBaseCases creates test cases from the base cases with the given options
func GenerateBaseCases(opts BaseCaseOptions) []rule_tester.InvalidTestCase {
	if opts.MutateCode == nil {
		opts.MutateCode = Identity
	}
	if opts.MutateDeclaration == nil {
		opts.MutateDeclaration = Identity
	}
	if opts.MutateOutput == nil {
		opts.MutateOutput = opts.MutateCode
	}
	if opts.Options == nil {
		opts.Options = map[string]any{
			"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
		}
	}

	var result []rule_tester.InvalidTestCase
	for _, bc := range RawBaseCases() {
		if opts.SkipIDs[bc.ID] {
			continue
		}

		// Replace operator placeholder
		chain := strings.ReplaceAll(bc.Chain, "${operator}", opts.Operator)
		outputChain := bc.OutputChain

		// Apply mutations
		declaration := opts.MutateDeclaration(bc.Declaration)
		code := opts.MutateCode(chain)
		output := opts.MutateOutput(outputChain)

		// Build full code with comment and declaration
		fullCode := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, code)
		fullOutput := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, output)

		tc := rule_tester.InvalidTestCase{
			Code:    fullCode,
			Options: opts.Options,
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		}

		if opts.UseSuggestionFixer {
			tc.Output = nil
			// Note: Suggestions not yet implemented in rule_tester
		} else {
			tc.Output = []string{fullOutput}
		}

		result = append(result, tc)
	}

	return result
}

// GenerateValidBaseCases creates valid test cases (no fix expected) from base cases
func GenerateValidBaseCases(opts BaseCaseOptions) []rule_tester.ValidTestCase {
	if opts.MutateCode == nil {
		opts.MutateCode = Identity
	}
	if opts.MutateDeclaration == nil {
		opts.MutateDeclaration = Identity
	}
	if opts.Options == nil {
		opts.Options = map[string]any{}
	}

	var result []rule_tester.ValidTestCase
	for _, bc := range RawBaseCases() {
		if opts.SkipIDs[bc.ID] {
			continue
		}

		// Replace operator placeholder
		chain := strings.ReplaceAll(bc.Chain, "${operator}", opts.Operator)

		// Apply mutations
		declaration := opts.MutateDeclaration(bc.Declaration)
		code := opts.MutateCode(chain)

		// Build full code with comment and declaration
		fullCode := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, code)

		result = append(result, rule_tester.ValidTestCase{
			Code:    fullCode,
			Options: opts.Options,
		})
	}

	return result
}

// Common mutation functions matching upstream

// ReplaceOperatorWithNotEqualNull replaces && with != null &&
func ReplaceOperatorWithNotEqualNull(s string) string {
	return strings.ReplaceAll(s, "&&", "!= null &&")
}

// ReplaceOperatorWithNotEqualUndefined replaces && with != undefined &&
func ReplaceOperatorWithNotEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "&&", "!= undefined &&")
}

// ReplaceOperatorWithStrictNotEqualNull replaces && with !== null &&
func ReplaceOperatorWithStrictNotEqualNull(s string) string {
	return strings.ReplaceAll(s, "&&", "!== null &&")
}

// ReplaceOperatorWithStrictNotEqualUndefined replaces && with !== undefined &&
func ReplaceOperatorWithStrictNotEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "&&", "!== undefined &&")
}

// RemoveNullFromType removes | null from type declarations
func RemoveNullFromType(s string) string {
	return strings.ReplaceAll(s, "| null", "")
}

// RemoveUndefinedFromType removes | undefined from type declarations
func RemoveUndefinedFromType(s string) string {
	return strings.ReplaceAll(s, "| undefined", "")
}

// AddTrailingAnd appends && bing to the chain
func AddTrailingAnd(s string) string {
	return strings.Replace(s, ";", " && bing;", 1)
}

// AddTrailingAndBingBong appends && bing.bong to the chain
func AddTrailingAndBingBong(s string) string {
	return strings.Replace(s, ";", " && bing.bong;", 1)
}

// OR operator mutations

// ReplaceOperatorWithEqualNull replaces || with == null ||
func ReplaceOperatorWithEqualNull(s string) string {
	return strings.ReplaceAll(s, "||", "== null ||")
}

// ReplaceOperatorWithEqualUndefined replaces || with == undefined ||
func ReplaceOperatorWithEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "||", "== undefined ||")
}

// ReplaceOperatorWithStrictEqualNull replaces || with === null ||
func ReplaceOperatorWithStrictEqualNull(s string) string {
	return strings.ReplaceAll(s, "||", "=== null ||")
}

// ReplaceOperatorWithStrictEqualUndefined replaces || with === undefined ||
func ReplaceOperatorWithStrictEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "||", "=== undefined ||")
}

// NegateExpression adds ! prefix to negate the expression for OR chains
func NegateExpression(s string) string {
	// This is simplified - upstream has more complex logic
	return "!" + s
}
