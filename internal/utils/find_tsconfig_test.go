package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
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
		expectedConfigs map[string]string
	}{
		{
			name:  "single file",
			files: []string{"file.ts"},
			expectedConfigs: map[string]string{
				filepath.Join(rootDir, "file.ts"): expectedConfigPath,
			},
		},
		{
			name:  "multiple files",
			files: []string{"file.ts", "foo.ts", "class.ts", "deprecated.ts"},
			expectedConfigs: map[string]string{
				filepath.Join(rootDir, "file.ts"):       expectedConfigPath,
				filepath.Join(rootDir, "foo.ts"):        expectedConfigPath,
				filepath.Join(rootDir, "class.ts"):      expectedConfigPath,
				filepath.Join(rootDir, "deprecated.ts"): expectedConfigPath,
			},
		},
		{
			name:            "empty file list",
			files:           []string{},
			expectedConfigs: map[string]string{},
		},
		{
			name:  "non-existent file",
			files: []string{"nonexistent.ts"},
			expectedConfigs: map[string]string{
				filepath.Join(rootDir, "nonexistent.ts"): "",
			},
		},
		{
			name:  "mixed existing and non-existing files",
			files: []string{"file.ts", "nonexistent.ts", "foo.ts"},
			expectedConfigs: map[string]string{
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

			assert.Equal(t, len(tc.expectedConfigs), len(results),
				"Number of results should match expected")

			for filePath, expectedConfig := range tc.expectedConfigs {
				res, exists := results[filePath]
				assert.Assert(t, exists, "Result should exist for %s", filePath)
				assert.Equal(t, expectedConfig, res.Config,
					"Config path should match for %s", filePath)

				if res.Config != "" {
					_, err := os.Stat(res.Config)
					assert.NilError(t, err, "Config file should exist: %s", res.Config)
				}
			}
		})
	}
}

// TestFindTsConfigParallel_Consistency verifies that the parallel
// implementation produces identical Config results to the sequential implementation
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
		res, exists := parallelResults[file]
		assert.Assert(t, exists, "Parallel should have result for %s", file)
		assert.Equal(t, expectedConfig, res.Config,
			"Parallel result should match sequential for %s", file)
	}
}

func TestFindTsConfigParallel_NearestConfig(t *testing.T) {
	dir := t.TempDir()

	tsconfigContent := `{
  "compilerOptions": { "lib": ["esnext"], "target": "esnext", "noEmit": true },
  "include": ["src"]
}`
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(tsconfigContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "main.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "spec.ts"), []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tsconfigPath := tspath.NormalizeSlashes(filepath.Join(dir, "tsconfig.json"))
	matchedFile := tspath.NormalizeSlashes(filepath.Join(dir, "src", "main.ts"))
	unmatchedFile := tspath.NormalizeSlashes(filepath.Join(dir, "tests", "spec.ts"))

	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, dir)
	results := resolver.FindTsConfigParallel([]string{matchedFile, unmatchedFile})

	matched, ok := results[matchedFile]
	assert.Assert(t, ok)
	assert.Equal(t, tsconfigPath, matched.Config)
	assert.Equal(t, tsconfigPath, matched.NearestConfig)

	// tests/spec.ts is outside include: Config empty, NearestConfig set.
	unmatched, ok := results[unmatchedFile]
	assert.Assert(t, ok)
	assert.Equal(t, "", unmatched.Config)
	assert.Equal(t, tsconfigPath, unmatched.NearestConfig)
}

func TestFindTsConfigParallel_NoNearestConfig(t *testing.T) {
	dir := t.TempDir()

	tsFile := filepath.Join(dir, "standalone.ts")
	if err := os.WriteFile(tsFile, []byte("export {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	normalizedFile := tspath.NormalizeSlashes(tsFile)
	fs := osvfs.FS()
	resolver := NewTsConfigResolver(fs, dir)
	results := resolver.FindTsConfigParallel([]string{normalizedFile})

	res, ok := results[normalizedFile]
	assert.Assert(t, ok)
	assert.Equal(t, "", res.Config)
	assert.Equal(t, "", res.NearestConfig)
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

	// Run 10 times and verify both Config and NearestConfig are stable.
	var firstConfigs, firstNearest map[string]string
	for i := range 10 {
		results := resolver.FindTsConfigParallel(testFiles)
		configs := make(map[string]string, len(results))
		nearest := make(map[string]string, len(results))
		for file, res := range results {
			configs[file] = res.Config
			nearest[file] = res.NearestConfig
		}

		if i == 0 {
			firstConfigs = configs
			firstNearest = nearest
		} else {
			assert.DeepEqual(t, firstConfigs, configs)
			assert.DeepEqual(t, firstNearest, nearest)
		}
	}
}
