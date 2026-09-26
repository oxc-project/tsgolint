package naming_convention

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

func TestNamingFormatUnicodeAndStrictHumps(t *testing.T) {
	cases := []struct {
		name, format string
		want         bool
	}{
		{"", "strictCamelCase", true},
		{"fooBar", "strictCamelCase", true},
		{"fooBAR", "strictCamelCase", false},
		{"FooBar", "StrictPascalCase", true},
		{"FOO", "StrictPascalCase", false},
		{"foo_bar", "snake_case", true},
		{"foo__bar", "snake_case", false},
		{"foo_", "snake_case", false},
		{"ß", "camelCase", true},
		{"ß", "PascalCase", false},
		{"ß", "UPPER_CASE", false},
		{"中文", "camelCase", true},
		{"中文", "PascalCase", true},
		{"𝔄", "camelCase", true}, // JS sees the leading high surrogate.
		{"𝔄", "PascalCase", true},
	}
	for _, tc := range cases {
		t.Run(tc.format+"/"+tc.name, func(t *testing.T) {
			if got := matchesNamingFormat(tc.name, tc.format); got != tc.want {
				t.Fatalf("matchesNamingFormat(%q, %q) = %v, want %v", tc.name, tc.format, got, tc.want)
			}
		})
	}
}

func TestNamingOptionNormalizationAndRegex(t *testing.T) {
	options := []any{
		map[string]any{"selector": "default", "format": []any{"camelCase"}},
		map[string]any{"selector": []any{"variable", "parameter"}, "format": nil, "filter": `^\u{2e}$`, "modifiers": []any{"unused"}},
		map[string]any{"selector": "variable", "format": []any{"UPPER_CASE"}, "filter": map[string]any{"regex": `^\p{Script=Greek}+$`, "match": true}, "types": []any{"string"}},
	}
	v := newNamingValidator(rule.RuleContext{}, options)
	if !v.needsModifier("unused") || v.needsModifier("exported") {
		t.Fatal("modifier demand was not derived from configs")
	}
	configs := v.configs["variable"]
	if len(configs) != 3 || configs[0].selector != namingSelectorBits["variable"] || configs[1].selector != namingSelectorBits["variable"] || configs[2].selector != -1 {
		t.Fatalf("wrong individual/meta selector precedence: %+v", configs)
	}
	if !configs[0].filter.test("Ω") || configs[0].filter.test("a") {
		t.Fatal("Unicode Script filter did not use JS property syntax")
	}
	if !configs[1].filter.test(".") || configs[1].filter.test("a") {
		t.Fatal("Unicode code-point escape must match a literal dot")
	}
	if configs[1].format != nil {
		t.Fatal("null format should retain other validation settings")
	}
	if got := namingUnicodeProperties(`\\p{Script=Greek}`); got != `\\p{Script=Greek}` {
		t.Fatalf("escaped backslash changed: %q", got)
	}
	greekAlias := namingRegex(`^\p{sc=Grek}$`, true)
	if !greekAlias.test("Ω") || greekAlias.test("A") {
		t.Fatal("script alias must match Greek letters")
	}
	// Valid regex config passes schema validation, so it must also compile in
	// regexp2 even when that engine lacks the ECMAScript property name.
	for _, tc := range []struct {
		pattern, text string
		want          bool
	}{
		{`^\p{ID_Start}$`, "A", true},
		{`^\p{IDS}$`, "Ω", true},
		{`^\p{ID_Start}$`, "1", false},
		{`^\p{Alphabetic}$`, "A", true},
		{`^\p{Alpha}$`, "😀", false},
		{`^\p{ASCII}$`, "A", true},
		{`^\p{ASCII}$`, "Ω", false},
		{`^\P{Any}$`, "A", false},
		{`^[x\P{Any}]$`, "x", true},
		{`^\p{XID_Continue}$`, "1", true},
		{`^\p{XIDC}$`, "😀", false},
		{`^\p{Script=Unknown}$`, "\U000E0000", true},
		{`^\p{sc=Zzzz}$`, "A", false},
		{`^\P{ID_Start}$`, "1", true},
		{`^\P{ID_Start}$`, "A", false},
		{`^[x\p{ID_Start}]$`, "Ω", true},
		{`^[x\P{ID_Start}]$`, "1", true},
		{`^[x\P{ID_Start}]$`, "Ω", false},
		{`^\p{ID_Start}$`, "𐐀", true},
		{`^\P{ID_Start}$`, "𐐀", false},
		{`^\p{gc=Lu}$`, "A", true},
		{`^\p{Uppercase_Letter}$`, "a", false},
		{`^\p{scx=Kana}$`, "ー", true},
		{`^\p{sc=Kana}$`, "ー", false},
		{`^[A\p{Script_Extensions=Katakana}]$`, "ー", true},
		{`^[A\p{scx=Kana}]$`, "A", true},
		{`^\P{scx=Kana}$`, "ー", false},
		{`^\P{scx=Kana}$`, "😀", true},
		{`^[X\P{scx=Kana}]$`, "ー", false},
		{`^[X\P{scx=Kana}]$`, "😀", true},
		{`^\p{scx=Kana}$`, "𚿰", true},
		{`^\P{scx=Kana}$`, "𚿰", false},
	} {
		if got := namingRegex(tc.pattern, true).test(tc.text); got != tc.want {
			t.Errorf("%q on %q: got %v, want %v", tc.pattern, tc.text, got, tc.want)
		}
	}
}

func TestNamingRegexMatchesLoneSurrogatesLikeJavaScript(t *testing.T) {
	// TypeScript encodes lone UTF-16 surrogates as their WTF-8 byte sequence.
	surrogate := string([]byte{0xed, 0xa0, 0x80}) // U+D800
	for _, tc := range []struct {
		pattern string
		text    string
		want    bool
	}{
		{`^\p{Surrogate}$`, surrogate, true},
		{`^\P{Surrogate}$`, surrogate, false},
		{`^\uD800$`, surrogate, true},
		{`^\p{Surrogate}$`, "𐐀", false}, // Supplementary characters are not surrogate code points.
		{`^\u{10400}$`, "𐐀", true},
		{`^A\p{Surrogate}Ω$`, "A" + surrogate + "Ω", true},
	} {
		if got := namingRegex(tc.pattern, true).test(tc.text); got != tc.want {
			t.Errorf("%q on %q: got %v, want %v", tc.pattern, tc.text, got, tc.want)
		}
	}
}

func TestNamingRegexDiagnosticString(t *testing.T) {
	if got := namingRegexString("a/b\n"); got != "/a\\/b\\n/u" {
		t.Fatalf("regex diagnostic string = %q", got)
	}
}

func TestNamingRegexpLiteralFlags(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"/a/mi", "/a/im"},
		{"/a/ygid", "/a/dgiy"},
		{"/a\\/b/sm", "/a\\/b/ms"},
	} {
		if got := namingRegexpLiteralString(tc.input); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNamingComputedKeyValues(t *testing.T) {
	for _, tc := range []struct {
		kind ast.Kind
		want string
	}{
		{ast.KindTrueKeyword, "true"},
		{ast.KindFalseKeyword, "false"},
		{ast.KindNullKeyword, "null"},
		{ast.KindNoSubstitutionTemplateLiteral, "undefined"},
		{ast.KindBinaryExpression, "undefined"},
	} {
		if got := namingNodeName(&ast.Node{Kind: tc.kind}); got != tc.want {
			t.Errorf("kind %v: name %q, want %q", tc.kind, got, tc.want)
		}
	}
}
