package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"gotest.tools/v3/assert"
)

func TestFindTsconfigForFile(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, rootDir)

	expectedConfigPath := filepath.Join(rootDir, "tsconfig.json")

	tests := []struct {
		name           string
		fileName       string
		expectedConfig string
		expectedFound  bool
	}{
		{
			name:           "existing file - file.ts",
			fileName:       "file.ts",
			expectedConfig: expectedConfigPath,
			expectedFound:  true,
		},
		{
			name:           "existing file - foo.ts",
			fileName:       "foo.ts",
			expectedConfig: expectedConfigPath,
			expectedFound:  true,
		},
		{
			name:           "existing file - class.ts",
			fileName:       "class.ts",
			expectedConfig: expectedConfigPath,
			expectedFound:  true,
		},
		{
			name:           "existing file - deprecated.ts",
			fileName:       "deprecated.ts",
			expectedConfig: expectedConfigPath,
			expectedFound:  true,
		},
		{
			name:           "non-existent file returns not found",
			fileName:       "nonexistent.ts",
			expectedConfig: "",
			expectedFound:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(rootDir, tc.fileName)
			config, found := resolver.FindTsconfigForFile(filePath, false)

			assert.Equal(t, tc.expectedFound, found,
				"Found flag should be %v for %s", tc.expectedFound, tc.fileName)
			assert.Equal(t, tc.expectedConfig, config,
				"Config path should be %s for %s", tc.expectedConfig, tc.fileName)

			if found && config != "" {
				_, err := os.Stat(config)
				assert.NilError(t, err, "Config file should exist: %s", config)
			}
		})
	}
}

func TestFindTsConfigParallel(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, rootDir)

	expectedConfigPath := filepath.Join(rootDir, "tsconfig.json")

	tests := []struct {
		name            string
		files           []string
		expectedResults map[string]string
	}{
		{
			name:  "single file",
			files: []string{"file.ts"},
			expectedResults: map[string]string{
				filepath.Join(rootDir, "file.ts"): expectedConfigPath,
			},
		},
		{
			name:  "multiple files",
			files: []string{"file.ts", "foo.ts", "class.ts", "deprecated.ts"},
			expectedResults: map[string]string{
				filepath.Join(rootDir, "file.ts"):       expectedConfigPath,
				filepath.Join(rootDir, "foo.ts"):        expectedConfigPath,
				filepath.Join(rootDir, "class.ts"):      expectedConfigPath,
				filepath.Join(rootDir, "deprecated.ts"): expectedConfigPath,
			},
		},
		{
			name:            "empty file list",
			files:           []string{},
			expectedResults: map[string]string{},
		},
		{
			name:  "non-existent file",
			files: []string{"nonexistent.ts"},
			expectedResults: map[string]string{
				filepath.Join(rootDir, "nonexistent.ts"): "",
			},
		},
		{
			name:  "mixed existing and non-existing files",
			files: []string{"file.ts", "nonexistent.ts", "foo.ts"},
			expectedResults: map[string]string{
				filepath.Join(rootDir, "file.ts"):        expectedConfigPath,
				filepath.Join(rootDir, "nonexistent.ts"): "",
				filepath.Join(rootDir, "foo.ts"):         expectedConfigPath,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to absolute paths
			filePaths := make([]string, len(tc.files))
			for i, file := range tc.files {
				filePaths[i] = filepath.Join(rootDir, file)
			}

			results := resolver.FindTsConfigParallel(filePaths)

			assert.Equal(t, len(tc.expectedResults), len(results),
				"Number of results should match expected")

			for filePath, expectedConfig := range tc.expectedResults {
				actualConfig, exists := results[filePath]
				assert.Assert(t, exists, "Result should exist for %s", filePath)
				assert.Equal(t, expectedConfig, actualConfig,
					"Config path should match for %s", filePath)

				if actualConfig != "" {
					_, err := os.Stat(actualConfig)
					assert.NilError(t, err, "Config file should exist: %s", actualConfig)
				}
			}
		})
	}
}

// TestFindTsConfigParallel_Consistency verifies that the parallel
// implementation produces identical results to the sequential implementation
func TestFindTsConfigParallel_Consistency(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, rootDir)

	testFiles := []string{
		filepath.Join(rootDir, "file.ts"),
		filepath.Join(rootDir, "foo.ts"),
		filepath.Join(rootDir, "class.ts"),
		filepath.Join(rootDir, "deprecated.ts"),
		filepath.Join(rootDir, "nonexistent.ts"),
	}

	// Get sequential results
	sequentialResults := make(map[string]string)
	for _, file := range testFiles {
		config, _ := resolver.FindTsconfigForFile(file, false)
		sequentialResults[file] = config
	}

	// Get parallel results
	parallelResults := resolver.FindTsConfigParallel(testFiles)

	// Verify consistency
	assert.Equal(t, len(sequentialResults), len(parallelResults),
		"Result count should match between sequential and parallel")

	for file, expectedConfig := range sequentialResults {
		actualConfig, exists := parallelResults[file]
		assert.Assert(t, exists, "Parallel should have result for %s", file)
		assert.Equal(t, expectedConfig, actualConfig,
			"Parallel result should match sequential for %s", file)
	}
}

// TestFindTsConfigParallel_Determinism ensures parallel execution
// produces consistent results across multiple runs
func TestFindTsConfigParallel_Determinism(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, rootDir)

	testFiles := []string{
		filepath.Join(rootDir, "file.ts"),
		filepath.Join(rootDir, "foo.ts"),
		filepath.Join(rootDir, "class.ts"),
	}

	// Run 10 times
	var firstRun map[string]string
	for i := range 10 {
		results := resolver.FindTsConfigParallel(testFiles)

		if i == 0 {
			firstRun = results
		} else {
			assert.DeepEqual(t, firstRun, results)
		}
	}
}
