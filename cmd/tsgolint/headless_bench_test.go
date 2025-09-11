package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// cloneVueCore clones the vuejs/core repository at the specified commit
func cloneVueCore(b *testing.B) string {
	b.Helper()

	// Create a temporary directory for the benchmark
	tmpDir := b.TempDir()
	repoPath := filepath.Join(tmpDir, "vue-core")

	// Clone the repository at the specific commit
	cmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/vuejs/core.git", repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to clone repository: %v\nOutput: %s", err, output)
	}

	// Checkout the specific commit
	cmd = exec.Command("git", "-C", repoPath, "fetch", "--depth", "1", "origin", "75220c7995a13a483ae9599a739075be1c8e17f8")
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to fetch commit: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "-C", repoPath, "checkout", "75220c7995a13a483ae9599a739075be1c8e17f8")
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to checkout commit: %v\nOutput: %s", err, output)
	}

	return repoPath
}

func findAllFilesForLinting(rootPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == "coverage" || name == "temp" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// BenchmarkHeadlessVueCore benchmarks running the headless linter on the entire Vue.js core repository
func BenchmarkHeadlessVueCore(b *testing.B) {
	// Clone the repository
	repoPath := cloneVueCore(b)

	// Find all TypeScript files
	files, err := findAllFilesForLinting(repoPath)
	if err != nil {
		b.Fatalf("Failed to find TypeScript files: %v", err)
	}

	if len(files) == 0 {
		b.Fatal("No TypeScript files found in the repository")
	}

	b.Logf("Found %d TypeScript files to lint", len(files))

	// Create the headless payload with all rules
	allRuleNames := make([]headlessRule, len(allRules))
	for i, r := range allRules {
		allRuleNames[i] = headlessRule{Name: r.Name}
	}

	payload := &headlessPayload{
		Version: 2,
		Configs: []headlessConfig{
			{
				FilePaths: files,
				Rules:     allRuleNames,
			},
		},
	}


	b.ReportAllocs()

	for b.Loop() {
		diagnosticCount := 0
		onDiagnostic := func(diagnostic rule.RuleDiagnostic) {
			diagnosticCount++
		}

		// Run the headless linter
		err := runHeadlessWithPayload(payload, repoPath, onDiagnostic)
		if err != nil {
			b.Fatalf("Linting failed: %v", err)
		}
	}
}
