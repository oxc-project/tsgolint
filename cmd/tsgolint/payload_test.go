package main

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"

	"gotest.tools/v3/assert"
)

func TestResolveConfigFiles(t *testing.T) {
	const cwd = "/repo"
	path := func(fileName string) tspath.Path {
		return tspath.ToPath(fileName, cwd, true)
	}
	configOf := func(filePaths ...string) headlessConfig {
		return headlessConfig{
			FilePaths: filePaths,
			Rules:     []headlessRule{{Name: "no-floating-promises"}},
		}
	}

	for _, testCase := range []struct {
		name    string
		configs []headlessConfig
		// expectedFiles are the files to lint, in the order they should be
		// handed to the tsconfig resolver.
		expectedFiles []string
	}{
		{
			name:          "absolute paths are kept",
			configs:       []headlessConfig{configOf("/repo/a.ts", "/repo/nested/b.ts")},
			expectedFiles: []string{"/repo/a.ts", "/repo/nested/b.ts"},
		},
		{
			name:          "a relative path is resolved against the working directory",
			configs:       []headlessConfig{configOf("a.ts", "./nested/b.ts")},
			expectedFiles: []string{"/repo/a.ts", "/repo/nested/b.ts"},
		},
		{
			name: "a file listed by several config groups is only resolved once",
			configs: []headlessConfig{
				configOf("/repo/a.ts"),
				configOf("a.ts", "/repo/b.ts"),
			},
			expectedFiles: []string{"/repo/a.ts", "/repo/b.ts"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resolved := resolveConfigFiles(testCase.configs, cwd, true)

			assert.Equal(t, len(resolved.normalizedFiles), len(testCase.expectedFiles), "unexpected resolved files")
			for i, fileName := range testCase.expectedFiles {
				assert.Equal(t, resolved.normalizedFiles[i], fileName, "unexpected resolved file at %d", i)
			}

			assert.Equal(t, len(resolved.fileConfigs), len(testCase.expectedFiles), "unexpected linted files")
			for _, fileName := range testCase.expectedFiles {
				rules, ok := resolved.fileConfigs[path(fileName)]
				assert.Assert(t, ok, "%s should be linted", fileName)
				assert.Equal(t, len(rules), 1, "%s should be linted with its config group's rules", fileName)
			}
		})
	}
}
