package tsgolint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule"
)

const (
	repoOrg       = "vuejs"
	repoName      = "core"
	vueCoreCommit = "75220c7995a13a483ae9599a739075be1c8e17f8"
)

func cloneRepo(b *testing.B, org, name, commit string) string {
	b.Helper()

	// Create a temporary directory for the benchmark
	tmpDir := b.TempDir()
	repoPath := filepath.Join(tmpDir, fmt.Sprintf("%s-%s", org, name))

	// Clone the repository at the specific commit
	cmd := exec.Command("git", "clone", "--depth", "1", fmt.Sprintf("https://github.com/%s/%s.git", org, name), repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to clone repository: %v\nOutput: %s", err, output)
	}

	// Checkout the specific commit
	cmd = exec.Command("git", "-C", repoPath, "fetch", "--depth", "1", "origin", commit)
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to fetch commit: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("git", "-C", repoPath, "checkout", commit)
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to checkout commit: %v\nOutput: %s", err, output)
	}

	return repoPath
}

func cloneVueCore(b *testing.B) string {
	b.Helper()

	vueCoreRepoPath := cloneRepo(b, repoOrg, repoName, vueCoreCommit)

	// Install dependencies using pnpm
	cmd := exec.Command("pnpm", "install", "--prefer-offline", "--no-frozen-lockfile")
	cmd.Dir = vueCoreRepoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("Failed to install dependencies: %v\nOutput: %s", err, output)
	}

	return vueCoreRepoPath
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

func BenchmarkHeadlessVueCore(b *testing.B) {
	// Use local e2e fixture files for benchmarking to avoid external dependencies
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	repoPath := filepath.Join(cwd, "../../e2e/fixtures")
	
	files, err := findAllFilesForLinting(repoPath)
	if err != nil {
		b.Fatalf("Failed to find TypeScript files: %v", err)
	}

	if len(files) == 0 {
		b.Fatal("No TypeScript files found in the repository")
	}

	b.Logf("Found %d TypeScript files to lint", len(files))

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

	for b.Loop() {
		diagnosticCount := 0
		onDiagnostic := func(diagnostic rule.RuleDiagnostic) {
			diagnosticCount++
		}

		err := runHeadlessWithPayload(payload, repoPath, onDiagnostic)
		if err != nil {
			b.Fatalf("Linting failed: %v", err)
		}
	}
}
