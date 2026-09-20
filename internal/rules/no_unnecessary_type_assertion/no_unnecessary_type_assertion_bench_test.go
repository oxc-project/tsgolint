package no_unnecessary_type_assertion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func incompatibleCallableBenchmarkCode(layerCount int, assertionCount int) string {
	var code strings.Builder
	code.WriteString("type Layer0<T> = T;\n")
	for layer := 1; layer <= layerCount; layer++ {
		fmt.Fprintf(&code, "type Layer%d<T> = (value: Layer%d<T>) => Layer%d<T>;\n", layer, layer-1, layer-1)
	}
	fmt.Fprintf(&code, `
type Watcher<T> = ((handler: Layer%d<T>) => void) & { readonly kind: 'watcher' };
type Subscription<T> = ((handler: T) => void) & { readonly kind: 'subscription' };

declare const watcher: Watcher<string>;
`, layerCount)
	for assertion := range assertionCount {
		fmt.Fprintf(&code, "const subscription%d = watcher as Subscription<string>;\n", assertion)
	}
	return code.String()
}

func BenchmarkNoUnnecessaryTypeAssertion(b *testing.B) {
	rule_tester.RunRuleBenchmark(
		fixtures.GetRootDir(),
		"tsconfig.minimal.json",
		b,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.BenchmarkTestCase{
			{
				Name: "optimization/incompatible-callables",
				// The old ordering recursively inspects every callable layer. The new ordering
				// rejects the differing `kind` property values during assignability instead.
				Code: incompatibleCallableBenchmarkCode(16, 32),
			},
			{
				Name: "tradeoff/nested-any",
				// This is the ordering counterexample from review: containsAny finds the nested
				// `any` cheaply, while the new ordering checks mutual assignability first.
				Code: `
type Nested<T> = (value: T) => T;
type Source<T> = (value: T) => T;
type Target<T> = (value: T) => T;

declare const source: Source<Nested<any>>;
const target = source as Target<Nested<any>>;
`,
			},
			{
				Name: "tradeoff/recursive-assignability",
				// These recursive generic callables require a structural assignability check.
				Code: `
type RecursiveSource<T> = (value: T) => RecursiveSource<{ value: T }>;
type RecursiveTarget<T> = (value: T) => RecursiveTarget<{ value: T }>;

declare const source: RecursiveSource<string>;
const target = source as RecursiveTarget<string>;
`,
			},
			{
				Name: "common/same-type",
				Code: `
declare const value: string;
const result = value as string;
`,
				ExpectedDiagnostics: 1,
			},
			{
				Name: "common/direct-any",
				Code: `
declare const value: any;
const result = value as string;
`,
			},
			{
				Name: "common/property-mismatch",
				Code: `
interface Source { shared: string; sourceOnly: number }
interface Target { shared: string; targetOnly: number }

declare const source: Source;
const target = source as Target;
`,
			},
			{
				Name: "common/type-argument-mismatch",
				Code: `
type Source<T> = { value: string };
type Target<T, U> = { value: string };

declare const source: Source<string>;
const target = source as Target<string, number>;
`,
			},
			{
				Name: "common/structurally-equivalent",
				Code: `
interface Source { value: string }
interface Target { value: string }

declare const source: Source;
const target = source as Target;
`,
				ExpectedDiagnostics: 1,
			},
			{
				Name: "common/structurally-incompatible",
				Code: `
interface Source { value: string }
interface Target { value: number }

declare const source: Source;
const target = source as Target;
`,
			},
		},
	)
}
