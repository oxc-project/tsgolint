package prefer_optional_chain

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
)

// BaseCase represents a single test case from upstream typescript-eslint
// Source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/prefer-optional-chain/base-cases.ts
type BaseCase struct {
	ID          int
	Chain       string
	Declaration string
	OutputChain string
}

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
		// case with a call expr inside the chain for some inefficient reason
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

type MutateFn func(string) string

func Identity(s string) string {
	return s
}

type BaseCaseOptions struct {
	Operator           string // "&&" or "||"
	MutateCode         MutateFn
	MutateDeclaration  MutateFn
	MutateOutput       MutateFn
	SkipIDs            map[int]bool
	UseSuggestionFixer bool
	Options            *PreferOptionalChainOptions
}

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
		if opts.UseSuggestionFixer {
			defaultOpts := rule_tester.OptionsFromJSON[PreferOptionalChainOptions](`{}`)
			opts.Options = &defaultOpts
		} else {
			defaultOpts := rule_tester.OptionsFromJSON[PreferOptionalChainOptions](`{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true}`)
			opts.Options = &defaultOpts
		}
	}

	var result []rule_tester.InvalidTestCase
	for _, bc := range RawBaseCases() {
		if opts.SkipIDs[bc.ID] {
			continue
		}

		chain := strings.ReplaceAll(bc.Chain, "${operator}", opts.Operator)
		outputChain := bc.OutputChain
		declaration := opts.MutateDeclaration(bc.Declaration)
		code := opts.MutateCode(chain)
		output := opts.MutateOutput(outputChain)

		fullCode := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, code)
		fullOutput := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, output)

		tc := rule_tester.InvalidTestCase{
			Code:    fullCode,
			Options: opts.Options,
		}

		if opts.UseSuggestionFixer {
			// No direct fix, but provide suggestion
			tc.Output = nil
			tc.Errors = []rule_tester.InvalidTestCaseError{{
				MessageId: "preferOptionalChain",
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "optionalChainSuggest",
					Output:    fullOutput,
				}},
			}}
		} else {
			tc.Output = []string{fullOutput}
			tc.Errors = []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}}
		}

		result = append(result, tc)
	}

	return result
}

func GenerateValidBaseCases(opts BaseCaseOptions) []rule_tester.ValidTestCase {
	if opts.MutateCode == nil {
		opts.MutateCode = Identity
	}
	if opts.MutateDeclaration == nil {
		opts.MutateDeclaration = Identity
	}
	if opts.Options == nil {
		defaultOpts := rule_tester.OptionsFromJSON[PreferOptionalChainOptions](`{}`)
		opts.Options = &defaultOpts
	}

	var result []rule_tester.ValidTestCase
	for _, bc := range RawBaseCases() {
		if opts.SkipIDs[bc.ID] {
			continue
		}

		chain := strings.ReplaceAll(bc.Chain, "${operator}", opts.Operator)
		declaration := opts.MutateDeclaration(bc.Declaration)
		code := opts.MutateCode(chain)
		fullCode := fmt.Sprintf("// %d\n%s\n%s", bc.ID, declaration, code)
		result = append(result, rule_tester.ValidTestCase{
			Code:    fullCode,
			Options: opts.Options,
		})
	}

	return result
}

// Common mutation functions matching upstream
func ReplaceOperatorWithNotEqualNull(s string) string {
	return strings.ReplaceAll(s, "&&", "!= null &&")
}
func ReplaceOperatorWithNotEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "&&", "!= undefined &&")
}
func ReplaceOperatorWithStrictNotEqualNull(s string) string {
	return strings.ReplaceAll(s, "&&", "!== null &&")
}
func ReplaceOperatorWithStrictNotEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "&&", "!== undefined &&")
}
func RemoveNullFromType(s string) string {
	return strings.ReplaceAll(s, "| null", "")
}
func RemoveUndefinedFromType(s string) string {
	return strings.ReplaceAll(s, "| undefined", "")
}
func AddTrailingAnd(s string) string {
	return strings.Replace(s, ";", " && bing;", 1)
}
func AddTrailingAndBingBong(s string) string {
	return strings.Replace(s, ";", " && bing.bong;", 1)
}

func ReplaceOperatorWithEqualNull(s string) string {
	return strings.ReplaceAll(s, "||", "== null ||")
}
func ReplaceOperatorWithEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "||", "== undefined ||")
}
func ReplaceOperatorWithStrictEqualNull(s string) string {
	return strings.ReplaceAll(s, "||", "=== null ||")
}
func ReplaceOperatorWithStrictEqualUndefined(s string) string {
	return strings.ReplaceAll(s, "||", "=== undefined ||")
}

func AddTrailingStrictEqualNull(fn MutateFn) MutateFn {
	return func(s string) string {
		s = fn(s)
		return strings.Replace(s, ";", " === null;", 1)
	}
}
func AddTrailingStrictEqualUndefined(fn MutateFn) MutateFn {
	return func(s string) string {
		s = fn(s)
		return strings.Replace(s, ";", " === undefined;", 1)
	}
}
func AddTrailingEqualNull(fn MutateFn) MutateFn {
	return func(s string) string {
		s = fn(s)
		return strings.Replace(s, ";", " == null;", 1)
	}
}
func AddTrailingEqualUndefined(fn MutateFn) MutateFn {
	return func(s string) string {
		s = fn(s)
		return strings.Replace(s, ";", " == undefined;", 1)
	}
}

func AddSpacingAfterDots(s string) string {
	return strings.ReplaceAll(s, ".", ".      ")
}
func AddNewlineAfterDots(s string) string {
	return strings.ReplaceAll(s, ".", ".\n")
}

var bracketContentRegex = regexp.MustCompile(`\[[^\]]+\]`)

func AddSpacingInsideBrackets(s string) string {
	return bracketContentRegex.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ReplaceAll(match, ".", ".      ")
	})
}
func AddNewlineInsideBrackets(s string) string {
	return bracketContentRegex.ReplaceAllStringFunc(s, func(match string) string {
		return strings.ReplaceAll(match, ".", ".\n")
	})
}
func DedupeInvalidTestCases(cases ...[]rule_tester.InvalidTestCase) []rule_tester.InvalidTestCase {
	seen := make(map[string]bool)
	var result []rule_tester.InvalidTestCase
	for _, batch := range cases {
		for _, tc := range batch {
			if !seen[tc.Code] {
				seen[tc.Code] = true
				result = append(result, tc)
			}
		}
	}
	return result
}
