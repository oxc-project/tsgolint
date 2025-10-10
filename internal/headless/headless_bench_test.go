package headless

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rules"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

var (
	commit = "03c265b1adee71ac88f833e065f7bb956b60550a"
	repo   = "https://github.com/microsoft/vscode.git"
)

// ensureVscodeRepo clones the microsoft/vscode repository at a specific commit
// (03c265b1...) into benchmarks/vscode if it doesn't already exist.
func ensureVscodeRepo(b *testing.B) string {
	cwd, err := os.Getwd()
	if err != nil { b.Fatalf("cwd: %v", err) }
	repoRoot := filepath.Dir(filepath.Dir(cwd))
	targetDir := filepath.Join(repoRoot, "benchmarks", "vscode")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
			b.Fatalf("mkdir: %v", err)
		}
		// Initialize and fetch the specific commit shallowly.
		if out, e := exec.Command("git", "init", targetDir).CombinedOutput(); e != nil {
			b.Fatalf("git init: %v\n%s", e, string(out))
		}
		fetch := exec.Command("git", "fetch", "--depth", "1", repo, commit)
		fetch.Dir = targetDir
		if out, e := fetch.CombinedOutput(); e != nil {
			b.Fatalf("git fetch: %v\n%s", e, string(out))
		}
		checkout := exec.Command("git", "checkout", "FETCH_HEAD")
		checkout.Dir = targetDir
		if out, e := checkout.CombinedOutput(); e != nil {
			b.Fatalf("git checkout: %v\n%s", e, string(out))
		}
	}
	return targetDir
}

// buildPayload builds the headless payload listing .ts / .tsx files under
// vscode/src/vs/editor (capped to a maximum for benchmark consistency).
func buildPayload(b *testing.B) []byte {
	repo := ensureVscodeRepo(b)
	editorDir := filepath.Join(repo, "src", "vs", "editor", "browser")
	var files []string
	filepath.WalkDir(editorDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }
		ext := filepath.Ext(path)
		if ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 { b.Fatalf("no files found in %s", editorDir) }
	type ruleEntry struct{ Name string `json:"name"` }
	payload := map[string]any{
		"version": 2,
		"configs": []any{ map[string]any{
			"file_paths": files,
			"rules": []ruleEntry{
				{Name: "no-floating-promises"},
				{Name: "no-unsafe-assignment"},
				{Name: "switch-exhaustiveness-check"},
			},
		}},
	}
	data, err := json.Marshal(payload)
	if err != nil { b.Fatalf("marshal payload: %v", err) }
	return data
}

func BenchmarkHeadless(b *testing.B) {
	// Skip running this unless explicitly requested.
	if os.Getenv("BENCH_HEADLESS") != "1" {
		b.Skip("set BENCH_HEADLESS=1 to run this benchmark (it requires cloning a large repo)")
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		p := buildPayload(b)
		r := bytes.NewReader(p)
		f, err := os.CreateTemp("", "tsgolint-bench-stdin-*.json")
		if err != nil { b.Fatalf("temp file: %v", err) }
		if _, err := f.Write(p); err != nil { b.Fatalf("write payload: %v", err) }
		if _, err := f.Seek(0, 0); err != nil { b.Fatalf("seek payload: %v", err) }

		os.Stdin = f
		// Redirect stdout to /dev/null during the timed section to avoid
		// counting large diagnostic output cost in allocations/time.
		devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil { b.Fatalf("open /dev/null: %v", err) }
		os.Stdout = devNull
		
		logLevel := utils.GetLogLevel()

		cwd, err := os.Getwd();
		if err != nil { b.Fatalf("error getting cwd: %v", err) }

		b.StartTimer()
		Run(p, rules.AllRulesByName, cwd, logLevel, os.Stdout)
		b.StopTimer()

		devNull.Close()
		f.Close()
		os.Remove(f.Name());
		_ = r
	}
}