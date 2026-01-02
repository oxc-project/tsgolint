package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"gotest.tools/v3/assert"
)

type ResolveResult struct {
	NormalizedPath string
	ConfigPath     string
	Found          bool
}

func toAbs(root, path string) string {
	return filepath.Join(root + path)
}

func generateTestFiles(n int) []string {
	base := []string{"foo.ts", "bar.ts", "baz.ts"}
	root := fixtures.GetRootDir()

	files := make([]string, n)
	for i := range n {
		b := base[i%len(base)]
		ext := path.Ext(b)
		stem := strings.TrimSuffix(b, ext)
		unique := fmt.Sprintf("%s__%06d%s", stem, i, ext)
		files[i] = toAbs(root, unique)
	}
	return files
}

func TestResolveManyCorrectness(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, rootDir)

	tests := []struct {
		name  string
		files []string
	}{
		{
			name:  "single file",
			files: []string{"file.ts"},
		},
		{
			name:  "multiple files",
			files: []string{"file.ts", "foo.ts", "class.ts", "deprecated.ts"},
		},
		{
			name:  "empty array",
			files: []string{},
		},
		{
			name:  "non-existent file",
			files: []string{"nonexistent.ts"},
		},
		{
			name:  "large batch",
			files: generateTestFiles(100),
		},
	}

	normalizeAll := func(files []string) []string {
		ret := make([]string, 0, len(files))
		for i := range files {
			abs := toAbs(rootDir, files[i])
			normalized := tspath.NormalizeSlashes(abs)
			ret = append(ret, normalized)
		}

		return ret

	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalizedPaths := normalizeAll(tc.files)

			// Check original version
			expectedResults := make([]ResolveResult, 0, len(tc.files))
			for i := range tc.files {
				normalized := normalizedPaths[i]
				config, found := resolver.FindTsconfigForFile(normalized, false)
				expectedResults = append(expectedResults, ResolveResult{
					NormalizedPath: normalized,
					ConfigPath:     config,
					Found:          found,
				})
			}

			// Check worker pool version
			actualResultsMap := resolver.FindTsConfigParallel(normalizedPaths)
			actualResults := make([]ResolveResult, 0, len(actualResultsMap))
			for k, v := range actualResultsMap {
				actualResults = append(actualResults, ResolveResult{
					NormalizedPath: k,
					ConfigPath:     v,
					Found:          v != "",
				})

			}

			sortResults := func(results []ResolveResult) {
				sort.Slice(results, func(i, j int) bool {
					return results[i].NormalizedPath < results[j].NormalizedPath
				})
			}
			sortResults(expectedResults)
			sortResults(actualResults)

			assert.Equal(t, len(expectedResults), len(actualResults), "result count mismatch")
			assert.DeepEqual(t, expectedResults, actualResults)
		})
	}
}
