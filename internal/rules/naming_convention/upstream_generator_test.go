package naming_convention

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

// These are the input templates passed to createTestCases.ts. Preserve the
// property order in Options: JSON.stringify(options) is part of each test's
// source code, so sorting those properties would change the case itself.
type upstreamTemplate struct {
	Code    []string       `json:"code"`
	Options jsontext.Value `json:"options"`
}

type upstreamTemplateFixture struct {
	Templates []upstreamTemplate `json:"templates"`
}

type upstreamFormatNames struct {
	Invalid []string `json:"invalid"`
	Valid   []string `json:"valid"`
}

type upstreamGeneratorFixture struct {
	FormatTestNames map[string]upstreamFormatNames `json:"formatTestNames"`
}

var upstreamFormatOrder = []string{
	"camelCase", "PascalCase", "snake_case", "strictCamelCase", "StrictPascalCase", "UPPER_CASE",
}

type upstreamVariant struct {
	name        string
	property    string
	value       any
	messageID   string
	messageData map[string]string
}

func generatedValidVariants(name string) []upstreamVariant {
	return []upstreamVariant{
		{name: name},
		{name: name, property: "leadingUnderscore", value: "forbid"},
		{name: "_" + name, property: "leadingUnderscore", value: "require"},
		{name: "__" + name, property: "leadingUnderscore", value: "requireDouble"},
		{name: "_" + name, property: "leadingUnderscore", value: "allow"},
		{name: name, property: "leadingUnderscore", value: "allow"},
		{name: "__" + name, property: "leadingUnderscore", value: "allowDouble"},
		{name: name, property: "leadingUnderscore", value: "allowDouble"},
		{name: "_" + name, property: "leadingUnderscore", value: "allowSingleOrDouble"},
		{name: name, property: "leadingUnderscore", value: "allowSingleOrDouble"},
		{name: "__" + name, property: "leadingUnderscore", value: "allowSingleOrDouble"},
		{name: name, property: "trailingUnderscore", value: "forbid"},
		{name: name + "_", property: "trailingUnderscore", value: "require"},
		{name: name + "__", property: "trailingUnderscore", value: "requireDouble"},
		{name: name + "_", property: "trailingUnderscore", value: "allow"},
		{name: name, property: "trailingUnderscore", value: "allow"},
		{name: name + "__", property: "trailingUnderscore", value: "allowDouble"},
		{name: name, property: "trailingUnderscore", value: "allowDouble"},
		{name: name + "_", property: "trailingUnderscore", value: "allowSingleOrDouble"},
		{name: name, property: "trailingUnderscore", value: "allowSingleOrDouble"},
		{name: name + "__", property: "trailingUnderscore", value: "allowSingleOrDouble"},
		{name: "MyPrefix" + name, property: "prefix", value: []string{"MyPrefix"}},
		{name: "MyPrefix2" + name, property: "prefix", value: []string{"MyPrefix1", "MyPrefix2"}},
		{name: name + "MySuffix", property: "suffix", value: []string{"MySuffix"}},
		{name: name + "MySuffix2", property: "suffix", value: []string{"MySuffix1", "MySuffix2"}},
	}
}

func generatedInvalidVariants(name, format string) []upstreamVariant {
	return []upstreamVariant{
		{name: name, messageID: "doesNotMatchFormat", messageData: map[string]string{"formats": format}},
		{name: "_" + name, property: "leadingUnderscore", value: "forbid", messageID: "unexpectedUnderscore", messageData: map[string]string{"position": "leading"}},
		{name: name, property: "leadingUnderscore", value: "require", messageID: "missingUnderscore", messageData: map[string]string{"count": "one", "position": "leading"}},
		{name: name, property: "leadingUnderscore", value: "requireDouble", messageID: "missingUnderscore", messageData: map[string]string{"count": "two", "position": "leading"}},
		{name: "_" + name, property: "leadingUnderscore", value: "requireDouble", messageID: "missingUnderscore", messageData: map[string]string{"count": "two", "position": "leading"}},
		{name: name + "_", property: "trailingUnderscore", value: "forbid", messageID: "unexpectedUnderscore", messageData: map[string]string{"position": "trailing"}},
		{name: name, property: "trailingUnderscore", value: "require", messageID: "missingUnderscore", messageData: map[string]string{"count": "one", "position": "trailing"}},
		{name: name, property: "trailingUnderscore", value: "requireDouble", messageID: "missingUnderscore", messageData: map[string]string{"count": "two", "position": "trailing"}},
		{name: name + "_", property: "trailingUnderscore", value: "requireDouble", messageID: "missingUnderscore", messageData: map[string]string{"count": "two", "position": "trailing"}},
		{name: name, property: "prefix", value: []string{"MyPrefix"}, messageID: "missingAffix", messageData: map[string]string{"affixes": "MyPrefix", "position": "prefix"}},
		{name: name, property: "prefix", value: []string{"MyPrefix1", "MyPrefix2"}, messageID: "missingAffix", messageData: map[string]string{"affixes": "MyPrefix1, MyPrefix2", "position": "prefix"}},
		{name: name, property: "suffix", value: []string{"MySuffix"}, messageID: "missingAffix", messageData: map[string]string{"affixes": "MySuffix", "position": "suffix"}},
		{name: name, property: "suffix", value: []string{"MySuffix1", "MySuffix2"}, messageID: "missingAffix", messageData: map[string]string{"affixes": "MySuffix1, MySuffix2", "position": "suffix"}},
	}
}

func readUpstreamCases(t *testing.T, file string) upstreamCases {
	t.Helper()
	var fixture upstreamTemplateFixture
	data := readGenerated[jsontext.Value](t, file)
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Templates) == 0 {
		var cases upstreamCases
		if err := json.Unmarshal(data, &cases, json.RejectUnknownMembers(true)); err != nil {
			t.Fatal(err)
		}
		return cases
	}
	if err := json.Unmarshal(data, &fixture, json.RejectUnknownMembers(true)); err != nil {
		t.Fatal(err)
	}
	var generator upstreamGeneratorFixture
	if err := json.Unmarshal(readGenerated[jsontext.Value](t, "generator.json"), &generator, json.RejectUnknownMembers(true)); err != nil {
		t.Fatal(err)
	}
	if len(generator.FormatTestNames) != len(upstreamFormatOrder) {
		t.Fatalf("unexpected format count: %d", len(generator.FormatTestNames))
	}
	var cases upstreamCases
	for _, template := range fixture.Templates {
		for _, format := range upstreamFormatOrder {
			names, ok := generator.FormatTestNames[format]
			if !ok || len(names.Valid) == 0 || len(names.Invalid) == 0 {
				t.Fatalf("missing generated names for %q", format)
			}
			for _, name := range names.Valid {
				for _, variant := range generatedValidVariants(name) {
					cases.Valid = append(cases.Valid, expandUpstreamCase(t, template, format, variant))
				}
			}
		}
	}
	for _, template := range fixture.Templates {
		for _, format := range upstreamFormatOrder {
			names := generator.FormatTestNames[format]
			for _, name := range names.Invalid {
				for _, variant := range generatedInvalidVariants(name, format) {
					cases.Invalid = append(cases.Invalid, expandUpstreamCase(t, template, format, variant))
				}
			}
		}
	}
	return cases
}

func expandUpstreamCase(t *testing.T, template upstreamTemplate, format string, variant upstreamVariant) jsontext.Value {
	t.Helper()
	ordered := bytes.Clone(template.Options)
	if err := (*jsontext.Value)(&ordered).Compact(); err != nil {
		t.Fatal(err)
	}
	if len(ordered) < 2 || ordered[0] != '{' || ordered[len(ordered)-1] != '}' {
		t.Fatalf("template options must be an object: %s", ordered)
	}
	formatJSON, err := json.Marshal([]string{format})
	if err != nil {
		t.Fatal(err)
	}
	ordered = append(ordered[:len(ordered)-1], []byte(`,"format":`)...)
	ordered = append(ordered, formatJSON...)
	if variant.property != "" {
		valueJSON, err := json.Marshal(variant.value)
		if err != nil {
			t.Fatal(err)
		}
		ordered = append(ordered, []byte(`,"`+variant.property+`":`)...)
		ordered = append(ordered, valueJSON...)
	}
	ordered = append(ordered, '}')
	lines := make([]string, len(template.Code))
	for i, code := range template.Code {
		lines[i] = strings.ReplaceAll(code, "%", variant.name)
	}
	caseValue := map[string]any{
		"code": "// " + string(ordered) + "\n" + strings.Join(lines, "\n"),
	}
	var options map[string]jsontext.Value
	if err := json.Unmarshal(ordered, &options); err != nil {
		t.Fatal(err)
	}
	options["filter"] = jsontext.Value(`{"match":false,"regex":".gnored"}`)
	caseValue["options"] = []map[string]jsontext.Value{options}
	if variant.messageID != "" {
		var selectorOption struct {
			Selector jsontext.Value `json:"selector"`
		}
		if err := json.Unmarshal(template.Options, &selectorOption); err != nil {
			t.Fatal(err)
		}
		var selectors []string
		if err := json.Unmarshal(selectorOption.Selector, &selectors); err != nil {
			var one string
			if err := json.Unmarshal(selectorOption.Selector, &one); err != nil {
				t.Fatal(err)
			}
			selectors = []string{one}
		}
		if len(selectors) == 0 || len(template.Code) == 0 {
			t.Fatal("generated invalid case has no selectors or code")
		}
		errors := make([]map[string]any, 0, len(selectors)*len(template.Code))
		for range template.Code {
			for _, selector := range selectors {
				errorValue := map[string]any{"messageId": variant.messageID}
				if !isMetaSelector(selector) {
					data := map[string]string{"name": variant.name, "type": selectorMessageType(selector)}
					for key, value := range variant.messageData {
						data[key] = value
					}
					errorValue["data"] = data
				}
				errors = append(errors, errorValue)
			}
		}
		caseValue["errors"] = errors
	}
	raw, err := json.Marshal(caseValue)
	if err != nil {
		t.Fatal(err)
	}
	if err := (*jsontext.Value)(&raw).Canonicalize(); err != nil {
		t.Fatal(fmt.Errorf("canonicalizing generated case: %w", err))
	}
	return raw
}

func isMetaSelector(selector string) bool {
	switch selector {
	case "default", "variableLike", "memberLike", "typeLike", "property", "method", "accessor":
		return true
	default:
		return false
	}
}

func selectorMessageType(selector string) string {
	var result strings.Builder
	for i, letter := range selector {
		if letter >= 'A' && letter <= 'Z' {
			result.WriteByte(' ')
		}
		if i == 0 && letter >= 'a' && letter <= 'z' {
			letter -= 'a' - 'A'
		}
		result.WriteRune(letter)
	}
	return result.String()
}
