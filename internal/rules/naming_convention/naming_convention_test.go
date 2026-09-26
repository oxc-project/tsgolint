package naming_convention

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

const upstreamCommit = "afb56de7d0123c79d2a2ffbb28ff7a8f8d951eab"

type upstreamManifest struct {
	Upstream    string            `json:"upstream"`
	Sources     map[string]string `json:"sources"`
	Suites      []upstreamSuite   `json:"suites"`
	Valid       int               `json:"valid"`
	Invalid     int               `json:"invalid"`
	Diagnostics int               `json:"diagnostics"`
}

type upstreamSuite struct {
	File    string   `json:"file"`
	Fixture string   `json:"fixture"`
	Valid   []string `json:"valid"`
	Invalid []string `json:"invalid"`
}

type upstreamCases struct {
	Valid   []jsontext.Value `json:"valid"`
	Invalid []jsontext.Value `json:"invalid"`
}

type upstreamCase struct {
	Code            string         `json:"code"`
	Options         jsontext.Value `json:"options"`
	LanguageOptions *struct {
		ParserOptions struct {
			Project         string `json:"project"`
			ProjectService  bool   `json:"projectService"`
			TSConfigRootDir string `json:"tsconfigRootDir"`
		} `json:"parserOptions"`
	} `json:"languageOptions"`
	Errors []struct {
		MessageID string            `json:"messageId"`
		Data      map[string]string `json:"data"`
		Line      int               `json:"line"`
		Column    int               `json:"column"`
		EndLine   int               `json:"endLine"`
		EndColumn int               `json:"endColumn"`
	} `json:"errors"`
}

var messagePlaceholder = regexp.MustCompile(`\{\{([A-Za-z]+)\}\}`)

func readGenerated[T any](t *testing.T, file string) T {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", file))
	if err != nil {
		t.Fatal(err)
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func loadUpstreamCase(t *testing.T, raw jsontext.Value) upstreamCase {
	t.Helper()
	var result upstreamCase
	if err := json.Unmarshal(raw, &result, json.RejectUnknownMembers(true)); err != nil {
		t.Fatal(err)
	}
	return result
}

func caseOptions(t *testing.T, raw jsontext.Value) any {
	t.Helper()
	if len(raw) == 0 {
		return nil // Upstream omitted options.
	}
	if bytes.Equal(raw, []byte("null")) {
		return jsontext.Value("null") // Preserve an explicit null if upstream adds one.
	}
	var options any
	if err := json.Unmarshal(raw, &options); err != nil {
		t.Fatal(err)
	}
	return options // In particular, [] remains an empty non-nil slice.
}

func caseTSConfig(t *testing.T, c upstreamCase) string {
	t.Helper()
	if c.LanguageOptions == nil {
		return ""
	}
	p := c.LanguageOptions.ParserOptions
	if p.Project != "./tsconfig.json" || p.ProjectService || p.TSConfigRootDir != "<fixtures>" {
		t.Fatalf("unknown upstream parser options: %+v", p)
	}
	return "tsconfig.naming-convention-project.json"
}

func expectedMessage(t *testing.T, templates map[string]string, id string, data map[string]string) string {
	t.Helper()
	// Upstream omits data for many generated cases. The RuleTester then checks
	// their message ID without asserting the interpolated description.
	if data == nil {
		return ""
	}
	message, ok := templates[id]
	if !ok {
		t.Fatalf("missing upstream message template %q", id)
	}
	used := map[string]bool{}
	message = messagePlaceholder.ReplaceAllStringFunc(message, func(placeholder string) string {
		name := placeholder[2 : len(placeholder)-2]
		used[name] = true
		value, ok := data[name]
		if !ok {
			t.Errorf("missing message data %q for %q", name, id)
		}
		return value
	})
	for name := range data {
		if !used[name] {
			t.Errorf("unused message data %q for %q", name, id)
		}
	}
	return message
}

// The JSON fixtures are the fully expanded upstream RuleTester cases. Running the
// generator with --check also compares every source hash against the pinned checkout.
func TestGeneratedUpstreamCaseParity(t *testing.T) {
	manifest := readGenerated[upstreamManifest](t, "manifest.json")
	if manifest.Upstream != upstreamCommit || len(manifest.Suites) != 17 || len(manifest.Sources) != 20 {
		t.Fatalf("unexpected upstream provenance: %s, %d suites, %d sources", manifest.Upstream, len(manifest.Suites), len(manifest.Sources))
	}
	if manifest.Valid != 8966 || manifest.Invalid != 7146 || manifest.Diagnostics != 44065 {
		t.Fatalf("unexpected upstream totals: %+v", manifest)
	}
	for name, digest := range manifest.Sources {
		if !strings.HasPrefix(name, "packages/eslint-plugin/") || len(digest) != 64 {
			t.Fatalf("invalid upstream source metadata %q: %q", name, digest)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			t.Fatal(err)
		}
	}
	files, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(manifest.Suites)+2 { // messages.json and manifest.json
		t.Fatalf("found %d generated files for %d suites", len(files), len(manifest.Suites))
	}
	seen := map[string]bool{}
	valid, invalid, diagnostics := 0, 0, 0
	for _, suite := range manifest.Suites {
		if seen[suite.Fixture] || !strings.HasSuffix(suite.File, ".test.ts") ||
			suite.Fixture != strings.TrimSuffix(filepath.Base(suite.File), ".test.ts")+".json" {
			t.Fatalf("invalid suite entry: %+v", suite)
		}
		seen[suite.Fixture] = true
		source := "packages/eslint-plugin/tests/rules/naming-convention/" + suite.File
		if manifest.Sources[source] == "" {
			t.Fatalf("missing source hash for %s", source)
		}
		cases := readGenerated[upstreamCases](t, suite.Fixture)
		if len(cases.Valid) != len(suite.Valid) || len(cases.Invalid) != len(suite.Invalid) {
			t.Fatalf("case count mismatch in %s", suite.Fixture)
		}
		for kind, items := range map[string][]jsontext.Value{"valid": cases.Valid, "invalid": cases.Invalid} {
			hashes := suite.Valid
			if kind == "invalid" {
				hashes = suite.Invalid
			}
			for i, raw := range items {
				compact := bytes.Clone(raw)
				value := jsontext.Value(compact)
				if err := value.Compact(); err != nil {
					t.Fatal(err)
				}
				digest := sha256.Sum256(value)
				if got := hex.EncodeToString(digest[:]); got != hashes[i] {
					t.Fatalf("case hash mismatch: %s/%s-%d: %s != %s", suite.Fixture, kind, i, got, hashes[i])
				}
				c := loadUpstreamCase(t, raw)
				if kind == "invalid" {
					if len(c.Errors) == 0 {
						t.Fatalf("invalid case without expected errors: %s/%d", suite.Fixture, i)
					}
					diagnostics += len(c.Errors)
				}
			}
		}
		valid += len(cases.Valid)
		invalid += len(cases.Invalid)
	}
	if valid != manifest.Valid || invalid != manifest.Invalid || diagnostics != manifest.Diagnostics {
		t.Fatalf("actual totals %d valid, %d invalid, %d diagnostics disagree with manifest", valid, invalid, diagnostics)
	}
	for _, file := range files {
		name := filepath.Base(file)
		if name != "manifest.json" && name != "messages.json" && !seen[name] {
			t.Fatalf("unlisted generated suite: %s", name)
		}
	}
}

func TestNamingConventionUpstream(t *testing.T) {
	manifest := readGenerated[upstreamManifest](t, "manifest.json")
	messages := readGenerated[map[string]string](t, "messages.json")
	for _, suite := range manifest.Suites {
		t.Run(strings.TrimSuffix(suite.Fixture, ".json"), func(t *testing.T) {
			cases := readGenerated[upstreamCases](t, suite.Fixture)
			valid := make([]rule_tester.ValidTestCase, 0, len(cases.Valid))
			invalid := make([]rule_tester.InvalidTestCase, 0, len(cases.Invalid))
			for _, raw := range cases.Valid {
				c := loadUpstreamCase(t, raw)
				if len(c.Errors) != 0 {
					t.Fatal("valid case has expected errors")
				}
				valid = append(valid, rule_tester.ValidTestCase{
					Code: c.Code, Options: caseOptions(t, c.Options), TSConfig: caseTSConfig(t, c),
				})
			}
			for _, raw := range cases.Invalid {
				c := loadUpstreamCase(t, raw)
				errors := make([]rule_tester.InvalidTestCaseError, 0, len(c.Errors))
				for _, expected := range c.Errors {
					errors = append(errors, rule_tester.InvalidTestCaseError{
						MessageId: expected.MessageID,
						Message:   expectedMessage(t, messages, expected.MessageID, expected.Data),
						Line:      expected.Line, Column: expected.Column,
						EndLine: expected.EndLine, EndColumn: expected.EndColumn,
					})
				}
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: c.Code, Options: caseOptions(t, c.Options), TSConfig: caseTSConfig(t, c), Errors: errors,
				})
			}
			if len(valid) != len(suite.Valid) || len(invalid) != len(suite.Invalid) {
				t.Fatal(fmt.Sprintf("loaded case count mismatch: %d valid, %d invalid", len(valid), len(invalid)))
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NamingConventionRule, valid, invalid)
		})
	}
}

func TestUpstreamOptionsShapes(t *testing.T) {
	if caseOptions(t, nil) != nil {
		t.Fatal("omitted options should be nil")
	}
	if got, ok := caseOptions(t, jsontext.Value("[]")).([]any); !ok || !slices.Equal(got, []any{}) {
		t.Fatalf("empty options should remain an empty slice: %#v", got)
	}
	if got, ok := caseOptions(t, jsontext.Value("null")).(jsontext.Value); !ok || string(got) != "null" {
		t.Fatalf("explicit null should remain distinguishable: %#v", got)
	}
}

func TestUpstreamMessageInterpolation(t *testing.T) {
	templates := map[string]string{"example": "{{type}} name `{{name}}`"}
	if got := expectedMessage(t, templates, "example", nil); got != "" {
		t.Fatalf("upstream error without data should assert only its ID, got %q", got)
	}
	if got := expectedMessage(t, templates, "example", map[string]string{"type": "Variable", "name": "bad_name"}); got != "Variable name `bad_name`" {
		t.Fatalf("unexpected interpolation: %q", got)
	}
}
