package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"gotest.tools/v3/assert"
)

type caseSensitivityFS struct {
	vfs.FS
	caseSensitive bool
}

func (fs caseSensitivityFS) UseCaseSensitiveFileNames() bool {
	return fs.caseSensitive
}

func TestFindTsConfigParallel_CaseInsensitivePaths(t *testing.T) {
	rootDir := t.TempDir()
	actualFile := filepath.Join(rootDir, "Included.ts")
	inputFile := filepath.Join(rootDir, "included.ts")
	configPath := filepath.Join(rootDir, "tsconfig.json")
	assert.NilError(t, os.WriteFile(actualFile, []byte("export const included = true;\n"), 0o644))
	assert.NilError(t, os.WriteFile(configPath, []byte(`{"include":["*.ts"]}`), 0o644))

	for _, tc := range []struct {
		name          string
		caseSensitive bool
		expected      string
	}{
		{name: "case insensitive", caseSensitive: false, expected: configPath},
		{name: "case sensitive", caseSensitive: true, expected: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := caseSensitivityFS{FS: osvfs.FS(), caseSensitive: tc.caseSensitive}
			resolver := NewTsConfigResolver(fs, rootDir)
			results := resolver.FindTsConfigParallel([]string{inputFile})
			assert.Equal(t, tc.expected, results[inputFile])
		})
	}
}

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

// Regression test for https://github.com/voidzero-dev/vite-plus/issues/1443.
func TestFindTsConfigParallel_UsesAncestorConfigWhenNearestConfigExcludesFile(t *testing.T) {
	rootDir := t.TempDir()
	filePath := filepath.Join(rootDir, "packages", "cli", "src", "utils", "terminal.ts")
	keptFilePath := filepath.Join(rootDir, "packages", "cli", "src", "utils", "kept.ts")
	nestedConfigPath := filepath.Join(rootDir, "packages", "cli", "tsconfig.json")
	rootConfigPath := filepath.Join(rootDir, "tsconfig.json")

	assert.NilError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	assert.NilError(t, os.WriteFile(filePath, []byte("import { styleText } from 'node:util';\n"), 0o644))
	assert.NilError(t, os.WriteFile(keptFilePath, []byte("export const kept = true;\n"), 0o644))
	assert.NilError(t, os.WriteFile(rootConfigPath, []byte(`{ "compilerOptions": { "types": ["node"] } }`), 0o644))
	assert.NilError(t, os.WriteFile(nestedConfigPath, []byte(`{
  "extends": "../../tsconfig.json",
  "files": [],
  "include": ["src/utils/kept.ts"],
  "exclude": []
}
`), 0o644))
	filePaths := []string{filePath, keptFilePath}
	for i := range 64 {
		anotherFile := filepath.Join(filepath.Dir(filePath), fmt.Sprintf("excluded-%d.ts", i))
		assert.NilError(t, os.WriteFile(anotherFile, []byte("export {};\n"), 0o644))
		filePaths = append(filePaths, anotherFile)
	}

	resolver := NewTsConfigResolver(osvfs.FS(), rootDir)

	config, found := resolver.FindTsconfigForFile(filePath, false)
	assert.Equal(t, true, found)
	assert.Equal(t, rootConfigPath, config)

	results := resolver.FindTsConfigParallel(filePaths)
	for _, file := range filePaths {
		if file != keptFilePath {
			assert.Equal(t, rootConfigPath, results[file])
		}
	}
	assert.Equal(t, nestedConfigPath, results[keptFilePath])
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
