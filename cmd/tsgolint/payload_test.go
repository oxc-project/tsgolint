package main

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"

	"gotest.tools/v3/assert"
)

const testCwd = "/repo"

// typeChecks and skipsTypeChecks are the `type_check` values a config group can
// carry; a nil `*bool` is a config group without the field.
var (
	typeChecks      = true
	skipsTypeChecks = false
)

func testPath(fileName string) tspath.Path {
	return tspath.ToPath(fileName, testCwd, true)
}

func configOf(typeCheck *bool, filePaths ...string) headlessConfig {
	return headlessConfig{
		FilePaths: filePaths,
		Rules:     []headlessRule{{Name: "no-floating-promises"}},
		TypeCheck: typeCheck,
	}
}

func TestDeserializePayloadTypeCheck(t *testing.T) {
	// `false` and "unset" must stay distinguishable on the wire.
	for _, testCase := range []struct {
		name    string
		payload string
		// expected is the `type_check` of every config group of the payload.
		expected []*bool
	}{
		{
			name:     "set to true",
			payload:  `{"version":2,"configs":[{"file_paths":["/a.ts"],"rules":[],"type_check":true}]}`,
			expected: []*bool{&typeChecks},
		},
		{
			name:     "set to false",
			payload:  `{"version":2,"configs":[{"file_paths":["/a.ts"],"rules":[],"type_check":false}]}`,
			expected: []*bool{&skipsTypeChecks},
		},
		{
			name:     "unset",
			payload:  `{"version":2,"configs":[{"file_paths":["/a.ts"],"rules":[]}]}`,
			expected: []*bool{nil},
		},
		{
			name: "one config group per state",
			payload: `{
				"version": 2,
				"configs": [
					{ "file_paths": ["/a.ts"], "rules": [], "type_check": true },
					{ "file_paths": ["/b.ts"], "rules": [], "type_check": false },
					{ "file_paths": ["/c.ts"], "rules": [] }
				],
				"report_semantic": true
			}`,
			expected: []*bool{&typeChecks, &skipsTypeChecks, nil},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			payload, err := deserializePayload([]byte(testCase.payload))
			assert.NilError(t, err, "couldn't deserialize payload")

			assert.Equal(t, len(payload.Configs), len(testCase.expected), "expected every config group")
			for i, expected := range testCase.expected {
				typeCheck := payload.Configs[i].TypeCheck
				if expected == nil {
					assert.Assert(t, typeCheck == nil, "config group %d should carry no `type_check`", i)
					continue
				}
				assert.Assert(t, typeCheck != nil, "config group %d should carry a `type_check`", i)
				assert.Equal(t, *typeCheck, *expected, "unexpected `type_check` for config group %d", i)
			}
		})
	}
}

func TestDeserializePayloadReportSemantic(t *testing.T) {
	payload, err := deserializePayload([]byte(`{
		"version": 2,
		"configs": [{ "file_paths": ["/a.ts"], "rules": [], "type_check": true }],
		"report_semantic": true
	}`))
	assert.NilError(t, err, "couldn't deserialize payload")

	assert.Equal(t, payload.ReportSemantic, true, "report_semantic should stay a payload-wide flag")
}

func TestDeserializePayloadV1TypeCheck(t *testing.T) {
	payload, err := deserializePayload([]byte(`{"files": [{"file_path": "/a.ts", "rules": ["no-floating-promises"]}]}`))
	assert.NilError(t, err, "couldn't deserialize payload")

	assert.Assert(t, payload.Configs[0].TypeCheck == nil, "V1 payloads never carry `type_check`")
}

func TestResolveConfigFiles(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		configs []headlessConfig
		// expectedFiles are the files to lint, in the order they should be
		// handed to the tsconfig resolver.
		expectedFiles []string
	}{
		{
			name:          "absolute paths are kept",
			configs:       []headlessConfig{configOf(nil, "/repo/a.ts", "/repo/nested/b.ts")},
			expectedFiles: []string{"/repo/a.ts", "/repo/nested/b.ts"},
		},
		{
			name:          "a relative path is resolved against the working directory",
			configs:       []headlessConfig{configOf(nil, "a.ts", "./nested/b.ts")},
			expectedFiles: []string{"/repo/a.ts", "/repo/nested/b.ts"},
		},
		{
			name: "a file listed by several config groups keeps its first position",
			configs: []headlessConfig{
				configOf(nil, "/repo/a.ts", "/repo/b.ts"),
				configOf(nil, "a.ts", "/repo/c.ts"),
			},
			expectedFiles: []string{"/repo/a.ts", "/repo/b.ts", "/repo/c.ts"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resolved := resolveConfigFiles(testCase.configs, testCwd, true)

			assert.Equal(t, len(resolved.normalizedFiles), len(testCase.expectedFiles), "unexpected resolved files")
			for i, fileName := range testCase.expectedFiles {
				assert.Equal(t, resolved.normalizedFiles[i], fileName, "unexpected resolved file at %d", i)
			}

			assert.Equal(t, len(resolved.fileConfigs), len(testCase.expectedFiles), "unexpected linted files")
			for _, fileName := range testCase.expectedFiles {
				rules, ok := resolved.fileConfigs[testPath(fileName)]
				assert.Assert(t, ok, "%s should be linted", fileName)
				assert.Equal(t, len(rules), 1, "%s should be linted with its config group's rules", fileName)
			}
		})
	}
}

// Rule 1 of the `type_check` semantics documented in payload.go: a payload
// where no config group carries the field keeps the historical behavior.
func TestTypeCheckRule1AbsentEverywhere(t *testing.T) {
	resolved := resolveConfigFiles([]headlessConfig{
		configOf(nil, "/repo/a.ts"),
		configOf(nil, "/repo/b.ts"),
	}, testCwd, true)

	assert.Assert(
		t,
		resolved.reportTypeErrorsForFile() == nil,
		"no file should be filtered out of the TypeScript diagnostics",
	)
	assert.Equal(t, len(resolved.skippedTypeCheckFiles), 0, "no file should be opted out")
}

// Rule 2 of the `type_check` semantics documented in payload.go: the last config
// group listing a file defines how it's linted, rules and `type_check` together.
func TestTypeCheckRule2ResolvedPerFile(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		configs []headlessConfig
		// expectedFiles maps every file the payload lints to whether it should
		// report the TypeScript diagnostics attached to it.
		expectedFiles map[string]bool
	}{
		{
			name:          "a single config group without the field",
			configs:       []headlessConfig{configOf(nil, "/repo/a.ts")},
			expectedFiles: map[string]bool{"/repo/a.ts": true},
		},
		{
			name:          "a single config group opting out",
			configs:       []headlessConfig{configOf(&skipsTypeChecks, "/repo/a.ts")},
			expectedFiles: map[string]bool{"/repo/a.ts": false},
		},
		{
			name: "config groups are independent",
			configs: []headlessConfig{
				configOf(&skipsTypeChecks, "/repo/a.ts"),
				configOf(nil, "/repo/b.ts"),
				configOf(&typeChecks, "/repo/c.ts"),
			},
			expectedFiles: map[string]bool{"/repo/a.ts": false, "/repo/b.ts": true, "/repo/c.ts": true},
		},
		{
			name: "a later config group without the field reports",
			configs: []headlessConfig{
				configOf(&skipsTypeChecks, "/repo/a.ts"),
				configOf(nil, "/repo/a.ts"),
			},
			expectedFiles: map[string]bool{"/repo/a.ts": true},
		},
		{
			name: "a later config group opting out wins over an unset one",
			configs: []headlessConfig{
				configOf(nil, "/repo/a.ts"),
				configOf(&skipsTypeChecks, "/repo/a.ts"),
			},
			expectedFiles: map[string]bool{"/repo/a.ts": false},
		},
		{
			name: "a later config group opting out wins over an opt-in",
			configs: []headlessConfig{
				configOf(&typeChecks, "/repo/a.ts"),
				configOf(&skipsTypeChecks, "/repo/a.ts"),
			},
			expectedFiles: map[string]bool{"/repo/a.ts": false},
		},
		{
			name: "a later config group opting in wins over an opt-out",
			configs: []headlessConfig{
				configOf(&skipsTypeChecks, "/repo/a.ts"),
				configOf(&typeChecks, "/repo/a.ts"),
			},
			expectedFiles: map[string]bool{"/repo/a.ts": true},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resolved := resolveConfigFiles(testCase.configs, testCwd, true)

			// The callback returned for the linter looks the source files up in
			// this set; the linter tests cover the callback itself.
			expectsFilter := false
			for _, typeChecked := range testCase.expectedFiles {
				expectsFilter = expectsFilter || !typeChecked
			}
			assert.Equal(t, resolved.reportTypeErrorsForFile() != nil, expectsFilter, "unexpected file filtering")

			for fileName, typeChecked := range testCase.expectedFiles {
				_, ok := resolved.fileConfigs[testPath(fileName)]
				assert.Assert(t, ok, "%s should be linted", fileName)

				_, skipped := resolved.skippedTypeCheckFiles[testPath(fileName)]
				assert.Equal(t, !skipped, typeChecked, "unexpected resolved `type_check` for %s", fileName)
			}
		})
	}
}

// Rule 2 of the `type_check` semantics documented in payload.go, rules half: a
// file listed by several config groups is linted with the last group's rules.
func TestTypeCheckRule2RulesOfTheLastConfigGroup(t *testing.T) {
	lastRules := []headlessRule{{Name: "no-unsafe-argument"}, {Name: "no-unsafe-call"}}
	resolved := resolveConfigFiles([]headlessConfig{
		configOf(nil, "/repo/a.ts", "/repo/b.ts"),
		{FilePaths: []string{"a.ts"}, Rules: lastRules, TypeCheck: &skipsTypeChecks},
	}, testCwd, true)

	rules := resolved.fileConfigs[testPath("/repo/a.ts")]
	assert.Equal(t, len(rules), len(lastRules), "/repo/a.ts should be linted with the last config group's rules")
	for i, rule := range lastRules {
		assert.Equal(t, rules[i].Name, rule.Name, "unexpected rule at %d", i)
	}

	// The config group that didn't list it again keeps its own rules.
	assert.Equal(t, len(resolved.fileConfigs[testPath("/repo/b.ts")]), 1, "/repo/b.ts should keep its rules")
}

// Rule 5 of the `type_check` semantics documented in payload.go: a config group
// listing no file has no effect. The path resolution the rule also covers is
// tested by TestResolveConfigFiles and the overlay FS tests.
func TestTypeCheckRule5ConfigGroupWithoutFiles(t *testing.T) {
	resolved := resolveConfigFiles([]headlessConfig{
		configOf(&skipsTypeChecks),
		configOf(&typeChecks),
		configOf(nil, "/repo/a.ts"),
	}, testCwd, true)

	assert.Equal(t, len(resolved.fileConfigs), 1, "only the config group with files should lint")
	assert.Assert(
		t,
		resolved.reportTypeErrorsForFile() == nil,
		"a config group without files shouldn't opt anything out",
	)
	assert.Equal(t, len(resolved.skippedTypeCheckFiles), 0, "no file should be opted out")
}

func TestResolveConfigFilesListedBySeveralConfigGroups(t *testing.T) {
	// A file taken over by another config group is recorded once, regardless
	// of its spelling; a file listed twice by one config group is not recorded.
	resolved := resolveConfigFiles([]headlessConfig{
		configOf(nil, "/repo/a.ts", "/repo/b.ts", "b.ts"),
		configOf(&skipsTypeChecks, "a.ts"),
		configOf(nil, "./a.ts", "/repo/c.ts"),
	}, testCwd, true)

	assert.Equal(t, len(resolved.filesInSeveralConfigs), 1, "only /repo/a.ts is taken over")
	assert.Equal(t, resolved.filesInSeveralConfigs[testPath("/repo/a.ts")], "/repo/a.ts", "unexpected recorded file")
	assert.Equal(t, len(resolved.normalizedFiles), 3, "every file should be linted once")
}
