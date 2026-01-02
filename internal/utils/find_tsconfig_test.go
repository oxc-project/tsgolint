package utils

import (
	"fmt"
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestResolveManyCorrectness(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.NewOSVFS()
	resolver := NewTsConfigResolver(fs, rootDir)

	testCases := []struct {
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
			name:  "duplicate files",
			files: []string{"file.ts", "file.ts", "foo.ts"},
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sequential execution (ground truth)
			expectedResults := make([]ResolveResult, 0, len(tc.files))
			for _, file := range tc.files {
				normalized := tspath.NormalizeSlashes(file)
				config, found := resolver.FindTsconfigForFile(normalized, false)
				expectedResults = append(expectedResults, ResolveResult{
					OriginalPath:   file,
					NormalizedPath: normalized,
					ConfigPath:     config,
					Found:          found,
				})
			}

			// Parallel execution (new implementation)
			actualResults := resolver.ResolveMany(tc.files, false)

			// Sort both results for comparison (order is not guaranteed)
			sortResults := func(results []ResolveResult) {
				sort.Slice(results, func(i, j int) bool {
					return results[i].OriginalPath < results[j].OriginalPath
				})
			}
			sortResults(expectedResults)
			sortResults(actualResults)

			// Compare results
			assert.Equal(t, len(expectedResults), len(actualResults), "result count mismatch")
			assert.DeepEqual(t, expectedResults, actualResults)
		})
	}
}

func TestResolveManyProperties(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.NewOSVFS()
	resolver := NewTsConfigResolver(fs, rootDir)
	files := []string{"file.ts", "foo.ts", "class.ts"}

	t.Run("idempotent", func(t *testing.T) {
		// Multiple executions should produce the same result
		result1 := resolver.ResolveMany(files, false)
		result2 := resolver.ResolveMany(files, false)

		sortResults := func(r []ResolveResult) {
			sort.Slice(r, func(i, j int) bool {
				return r[i].OriginalPath < r[j].OriginalPath
			})
		}
		sortResults(result1)
		sortResults(result2)

		assert.DeepEqual(t, result1, result2)
	})

	t.Run("all files in result", func(t *testing.T) {
		results := resolver.ResolveMany(files, false)
		assert.Equal(t, len(results), len(files))

		// All files should be in the result
		resultPaths := make(map[string]bool)
		for _, r := range results {
			resultPaths[r.OriginalPath] = true
		}
		for _, f := range files {
			assert.Assert(t, resultPaths[f], "file %s not in results", f)
		}
	})
}

// Lightweight race condition check - just run the normal test with -race flag
func TestResolveManyRace(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.NewOSVFS()
	resolver := NewTsConfigResolver(fs, rootDir)
	files := []string{"file.ts", "foo.ts", "class.ts", "deprecated.ts"}

	// Run multiple times to increase chance of detecting races
	for i := 0; i < 10; i++ {
		results := resolver.ResolveMany(files, false)
		assert.Equal(t, len(results), len(files))
	}
}

// Benchmark: Sequential vs Parallel
func BenchmarkResolveManyComparison(b *testing.B) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.NewOSVFS()
	resolver := NewTsConfigResolver(fs, rootDir)
	files := generateTestFiles(100)

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, file := range files {
				normalized := tspath.NormalizeSlashes(file)
				resolver.FindTsconfigForFile(normalized, false)
			}
		}
	})

	b.Run("Parallel_ResolveMany", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resolver.ResolveMany(files, false)
		}
	})
}

func BenchmarkResolveManyScalability(b *testing.B) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.NewOSVFS()
	resolver := NewTsConfigResolver(fs, rootDir)

	sizes := []int{10, 50, 100, 500}
	for _, size := range sizes {
		files := generateTestFiles(size)
		b.Run(fmt.Sprintf("Files_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				resolver.ResolveMany(files, false)
			}
		})
	}
}

// Helper function to generate test file paths
func generateTestFiles(n int) []string {
	// Use existing fixture files in a cycle
	baseFiles := []string{"file.ts", "foo.ts", "class.ts", "deprecated.ts"}
	files := make([]string, n)
	for i := 0; i < n; i++ {
		files[i] = baseFiles[i%len(baseFiles)]
	}
	return files
}
