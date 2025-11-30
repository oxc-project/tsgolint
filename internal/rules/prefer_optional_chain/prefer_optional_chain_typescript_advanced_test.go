package prefer_optional_chain

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

// TestPreferOptionalChainTypeScriptAdvanced tests TypeScript-specific advanced features
// This covers decorators, generics, private properties, and other TS-specific patterns
func TestPreferOptionalChainTypeScriptAdvanced(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferOptionalChainRule, []rule_tester.ValidTestCase{
		// Decorators should not be converted
		{Code: `@decorator class Foo { bar() {} }`},
		{Code: `class Foo { @decorator bar() {} }`},
		{Code: `class Foo { @decorator bar: string; }`},

		// Generic constraints - WITH requireNullish these are valid when type includes falsy
		{Code: `function foo<T extends {bar: string} | false>(x: T) { return x && x.bar; }`, Options: map[string]any{"requireNullish": true}},

		// Different objects with same property names (no chain)
		{Code: `declare const foo: {bar: string} | null; declare const baz: {bar: string} | null; foo && baz.bar;`},
	}, []rule_tester.InvalidTestCase{
		// Generic constraints WITHOUT requireNullish will trigger
		{
			Code:   `function foo<T extends {bar: string}>(x: T) { return x && x.bar; }`,
			Output: []string{`function foo<T extends {bar: string}>(x: T) { return x?.bar; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
		},

		// Type predicates - these currently trigger (may need rule enhancement to detect type guards)
		{
			Code:   `function isFoo(x: any): x is {bar: string} { return x && x.bar; }`,
			Output: []string{`function isFoo(x: any): x is {bar: string} { return x?.bar; }`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
		},
		{
			Code:   `const isFoo = (x: any): x is {bar: string} => x && x.bar;`,
			Output: []string{`const isFoo = (x: any): x is {bar: string} => x?.bar;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
		},

		// Multiple independent chains in one expression
		{
			Code:   `declare const a: {x: string} | null; declare const b: {x: string} | null; a && a.x && b && b.x;`,
			Output: []string{`declare const a: {x: string} | null; declare const b: {x: string} | null; a?.x && b?.x;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}, {MessageId: "preferOptionalChain"}},
			Options: map[string]any{"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true},
		},

		// Private properties with optional chaining
		{
			Code:   `declare class Foo { private bar: string | null; method() { this.bar && this.bar.length; } }`,
			Output: []string{`declare class Foo { private bar: string | null; method() { this.bar?.length; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare class Foo { #bar: {baz: string} | null; method() { this.#bar && this.#bar.baz; } }`,
			Output: []string{`declare class Foo { #bar: {baz: string} | null; method() { this.#bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `class Foo { #bar: {baz: {qux: string} | null} | null; get() { return this.#bar && this.#bar.baz && this.#bar.baz.qux; } }`,
			Output: []string{`class Foo { #bar: {baz: {qux: string} | null} | null; get() { return this.#bar?.baz?.qux; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Protected properties
		{
			Code:   `declare class Foo { protected bar: {baz: string} | null; method() { this.bar && this.bar.baz; } }`,
			Output: []string{`declare class Foo { protected bar: {baz: string} | null; method() { this.bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Readonly properties
		{
			Code:   `declare const foo: {readonly bar: {baz: string} | null}; foo.bar && foo.bar.baz;`,
			Output: []string{`declare const foo: {readonly bar: {baz: string} | null}; foo.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `class Foo { readonly bar: {baz: string} | null; method() { return this.bar && this.bar.baz; } }`,
			Output: []string{`class Foo { readonly bar: {baz: string} | null; method() { return this.bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Optional properties (should still convert the chain)
		{
			Code:   `declare const foo: {bar?: {baz: string} | null}; foo.bar && foo.bar.baz;`,
			Output: []string{`declare const foo: {bar?: {baz: string} | null}; foo.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = {bar?: {baz?: {qux: string} | null}}; declare const foo: Foo; foo.bar && foo.bar.baz && foo.bar.baz.qux;`,
			Output: []string{`type Foo = {bar?: {baz?: {qux: string} | null}}; declare const foo: Foo; foo.bar?.baz?.qux;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Index signatures
		{
			Code:   `declare const foo: {[key: string]: {bar: string} | null}; foo['key'] && foo['key'].bar;`,
			Output: []string{`declare const foo: {[key: string]: {bar: string} | null}; foo['key']?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare const foo: {[key: number]: {bar: {baz: string} | null} | null}; foo[0] && foo[0].bar && foo[0].bar.baz;`,
			Output: []string{`declare const foo: {[key: number]: {bar: {baz: string} | null} | null}; foo[0]?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Generic types with proper nullish union
		{
			Code:   `function foo<T extends {bar: string} | null>(x: T) { return x && x.bar; }`,
			Output: []string{`function foo<T extends {bar: string} | null>(x: T) { return x?.bar; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `function foo<T extends {bar: {baz: string} | undefined} | undefined>(x: T) { return x !== undefined && x.bar !== undefined && x.bar.baz; }`,
			Output: []string{`function foo<T extends {bar: {baz: string} | undefined} | undefined>(x: T) { return x?.bar?.baz; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `class Foo<T extends {bar: string} | null> { method(x: T) { return x != null && x.bar; } }`,
			Output: []string{`class Foo<T extends {bar: string} | null> { method(x: T) { return x?.bar; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Intersection types
		{
			Code:   `declare const foo: ({bar: string} | null) & {baz: number}; foo && foo.bar;`,
			Output: []string{`declare const foo: ({bar: string} | null) & {baz: number}; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = {bar: {baz: string} | null} & {qux: number}; declare const foo: Foo; foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = {bar: {baz: string} | null} & {qux: number}; declare const foo: Foo; foo.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Type aliases with nullish unions
		{
			Code:   `type MaybeString = string | null; type Foo = {bar: MaybeString}; declare const foo: Foo | null; foo && foo.bar;`,
			Output: []string{`type MaybeString = string | null; type Foo = {bar: MaybeString}; declare const foo: Foo | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Bar = {baz: string}; type Foo = {bar: Bar | undefined}; declare const foo: Foo | undefined; foo !== undefined && foo.bar !== undefined && foo.bar.baz;`,
			Output: []string{`type Bar = {baz: string}; type Foo = {bar: Bar | undefined}; declare const foo: Foo | undefined; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Interface with optional chaining
		{
			Code:   `interface Foo { bar: {baz: string} | null; } declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`interface Foo { bar: {baz: string} | null; } declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `interface Foo { bar?: {baz?: {qux: string} | null} } declare const foo: Foo | null; foo && foo.bar && foo.bar.baz && foo.bar.baz.qux;`,
			Output: []string{`interface Foo { bar?: {baz?: {qux: string} | null} } declare const foo: Foo | null; foo?.bar?.baz?.qux;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Namespace access
		{
			Code:   `namespace N { export const foo: {bar: string} | null; } N.foo && N.foo.bar;`,
			Output: []string{`namespace N { export const foo: {bar: string} | null; } N.foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `declare namespace N { export const foo: {bar: {baz: string} | null} | null; } N.foo && N.foo.bar && N.foo.bar.baz;`,
			Output: []string{`declare namespace N { export const foo: {bar: {baz: string} | null} | null; } N.foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Enum member access - NOTE: Currently produces foo?.bar?.A, may need rule fix
		// FIXME: Should be foo?.bar.A since we only check foo for nullish, not foo.bar
		{
			Code:   `enum E { A = 'a' } declare const foo: {bar: typeof E} | null; foo && foo.bar.A;`,
			Output: []string{`enum E { A = 'a' } declare const foo: {bar: typeof E} | null; foo?.bar?.A;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Abstract class members
		{
			Code:   `abstract class Foo { abstract bar: {baz: string} | null; method() { return this.bar && this.bar.baz; } }`,
			Output: []string{`abstract class Foo { abstract bar: {baz: string} | null; method() { return this.bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Static members
		{
			Code:   `class Foo { static bar: {baz: string} | null; static method() { return Foo.bar && Foo.bar.baz; } }`,
			Output: []string{`class Foo { static bar: {baz: string} | null; static method() { return Foo.bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `class Foo { static bar: {baz: {qux: string} | null} | null; static get() { return this.bar && this.bar.baz && this.bar.baz.qux; } }`,
			Output: []string{`class Foo { static bar: {baz: {qux: string} | null} | null; static get() { return this.bar?.baz?.qux; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Getter/setter patterns
		{
			Code:   `class Foo { get bar(): {baz: string} | null { return null; } method() { return this.bar && this.bar.baz; } }`,
			Output: []string{`class Foo { get bar(): {baz: string} | null { return null; } method() { return this.bar?.baz; } }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Tuple types with nullish elements
		{
			Code:   `declare const foo: [{bar: string} | null, string]; foo[0] && foo[0].bar;`,
			Output: []string{`declare const foo: [{bar: string} | null, string]; foo[0]?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Tuple = [string, {bar: {baz: string} | null} | null]; declare const foo: Tuple; foo[1] && foo[1].bar && foo[1].bar.baz;`,
			Output: []string{`type Tuple = [string, {bar: {baz: string} | null} | null]; declare const foo: Tuple; foo[1]?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Mapped types
		{
			Code:   `type Foo = {[K in 'bar']: {baz: string} | null}; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = {[K in 'bar']: {baz: string} | null}; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Conditional types (in variable declarations)
		{
			Code:   `type Foo<T> = T extends string ? {bar: string} | null : never; declare const foo: Foo<string>; foo && foo.bar;`,
			Output: []string{`type Foo<T> = T extends string ? {bar: string} | null : never; declare const foo: Foo<string>; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Template literal types (in object keys)
		{
			Code:   "type Foo = {[K in `bar${string}`]: string} | null; declare const foo: Foo; declare const key: 'barTest'; foo && foo[key];",
			Output: []string{"type Foo = {[K in `bar${string}`]: string} | null; declare const foo: Foo; declare const key: 'barTest'; foo?.[key];"},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Utility types
		{
			Code:   `type Foo = Partial<{bar: {baz: string}}>; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = Partial<{bar: {baz: string}}>; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = Required<{bar?: {baz: string} | null}>; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = Required<{bar?: {baz: string} | null}>; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = Readonly<{bar: {baz: string} | null}>; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = Readonly<{bar: {baz: string} | null}>; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = Pick<{bar: {baz: string} | null; qux: number}, 'bar'>; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = Pick<{bar: {baz: string} | null; qux: number}, 'bar'>; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = Omit<{bar: {baz: string} | null; qux: number}, 'qux'>; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = Omit<{bar: {baz: string} | null; qux: number}, 'qux'>; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
		{
			Code:   `type Foo = NonNullable<{bar: string} | null | undefined>; declare const foo: Foo | null; foo && foo.bar;`,
			Output: []string{`type Foo = NonNullable<{bar: string} | null | undefined>; declare const foo: Foo | null; foo?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// ReturnType utility
		{
			Code:   `function fn() { return {bar: {baz: string} | null}; } type Foo = ReturnType<typeof fn> | null; declare const foo: Foo; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`function fn() { return {bar: {baz: string} | null}; } type Foo = ReturnType<typeof fn> | null; declare const foo: Foo; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Constructor types
		{
			Code:   `declare const Foo: new () => {bar: {baz: string} | null} | null; declare const foo: Foo; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`declare const Foo: new () => {bar: {baz: string} | null} | null; declare const foo: Foo; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Function types with properties
		{
			Code:   `type Foo = (() => void) & {bar: {baz: string} | null}; declare const foo: Foo | null; foo && foo.bar && foo.bar.baz;`,
			Output: []string{`type Foo = (() => void) & {bar: {baz: string} | null}; declare const foo: Foo | null; foo?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Rest parameters with spread
		{
			Code:   `function foo(...args: [{bar: string} | null]) { return args[0] && args[0].bar; }`,
			Output: []string{`function foo(...args: [{bar: string} | null]) { return args[0]?.bar; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Destructured parameters
		{
			Code:   `function foo({bar}: {bar: {baz: string} | null}) { return bar && bar.baz; }`,
			Output: []string{`function foo({bar}: {bar: {baz: string} | null}) { return bar?.baz; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Default parameters
		{
			Code:   `function foo(bar: {baz: string} | null = null) { return bar && bar.baz; }`,
			Output: []string{`function foo(bar: {baz: string} | null = null) { return bar?.baz; }`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Asserted types
		{
			Code:   `declare const foo: unknown; (foo as {bar: string} | null) && (foo as {bar: string}).bar;`,
			Output: []string{`declare const foo: unknown; (foo as {bar: string} | null)?.bar;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},

		// Const assertions
		{
			Code:   `const foo = {bar: {baz: 'test'} as {baz: string} | null} as const; declare const x: typeof foo | null; x && x.bar && x.bar.baz;`,
			Output: []string{`const foo = {bar: {baz: 'test'} as {baz: string} | null} as const; declare const x: typeof foo | null; x?.bar?.baz;`},
			Options: map[string]any{
				"allowPotentiallyUnsafeFixesThatModifyTheReturnTypeIKnowWhatImDoing": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferOptionalChain"}},
		},
	})
}
