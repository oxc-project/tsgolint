package rule_tester

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

// snapshotDir is the directory where snapshot files are stored,
// computed once relative to this source file.
var snapshotDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	snapshotDir = filepath.Join(filepath.Dir(file), "__snapshots__")
}

// snapshotter matches test output against stored snapshot files.
type snapshotter struct {
	filename string
}

func newSnapshotter(filename string) *snapshotter {
	return &snapshotter{filename: filename}
}

// MatchSnapshot compares content against the stored snapshot for the current test.
// If the snapshot doesn't exist, it is created. If UPDATE_SNAPS=true, the snapshot
// is overwritten. Otherwise, a mismatch fails the test.
func (s *snapshotter) MatchSnapshot(t *testing.T, content string) {
	t.Helper()

	path := filepath.Join(snapshotDir, s.filename+".snap")
	key := fmt.Sprintf("[%s - 1]", t.Name())
	update := os.Getenv("UPDATE_SNAPS") == "true"

	snapshotRegistry.matchSnapshot(t, path, key, content, update)
}

// registry is the global snapshot file cache, ensuring each .snap file is
// loaded at most once and writes are serialized.
var snapshotRegistry = &snapRegistry{
	files: make(map[string]*snapshotFile),
}

type snapRegistry struct {
	mu    sync.Mutex
	files map[string]*snapshotFile
}

func (r *snapRegistry) getFile(path string) *snapshotFile {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sf, ok := r.files[path]; ok {
		return sf
	}

	sf := &snapshotFile{
		path:    path,
		entries: make(map[string]string),
	}
	r.files[path] = sf
	return sf
}

func (r *snapRegistry) matchSnapshot(t *testing.T, path, key, content string, update bool) {
	t.Helper()

	sf := r.getFile(path)
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if !sf.loaded {
		sf.load()
	}

	existing, exists := sf.entries[key]

	if update || !exists {
		sf.entries[key] = content
		sf.write(t)
		return
	}

	if existing != content {
		t.Errorf("Snapshot mismatch for %s.\nRun with UPDATE_SNAPS=true to update.\n\n--- Snapshot ---\n%s\n\n--- Actual ---\n%s", key, existing, content)
	}
}

type snapshotFile struct {
	mu      sync.Mutex
	path    string
	entries map[string]string
	loaded  bool
}

func (sf *snapshotFile) load() {
	sf.loaded = true

	data, err := os.ReadFile(sf.path)
	if err != nil {
		return
	}

	sf.entries = parseSnapshotFile(string(data))
}

func (sf *snapshotFile) write(t *testing.T) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(sf.path), 0o755); err != nil {
		t.Fatalf("could not create snapshot directory: %v", err)
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(sf.entries))
	for k := range sf.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&sb, "\n%s\n%s\n---\n", key, sf.entries[key])
	}

	if err := os.WriteFile(sf.path, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("could not write snapshot file: %v", err)
	}
}

// parseSnapshotFile parses a .snap file into a map of key -> content.
// Format: each entry is "\n[TestName - N]\ncontent\n---\n".
func parseSnapshotFile(data string) map[string]string {
	entries := make(map[string]string)

	blocks := strings.Split(data, "---\n")

	for _, block := range blocks {
		block = strings.TrimLeft(block, "\n")
		if block == "" {
			continue
		}

		newline := strings.IndexByte(block, '\n')
		if newline == -1 {
			continue
		}

		key := block[:newline]
		if !strings.HasPrefix(key, "[") || !strings.HasSuffix(key, "]") {
			continue
		}

		content := block[newline+1:]
		content = strings.TrimRight(content, "\n")

		entries[key] = content
	}

	return entries
}

// renderSourceAnnotation renders a source code snippet with a single underline annotation.
// marker is the character used for underlining (e.g. '~' or '^').
// label is an optional label appended after the underline.
func renderSourceAnnotation(code string, sourceFile *ast.SourceFile, textRange core.TextRange, marker byte, label string) string {
	if sourceFile == nil {
		return ""
	}

	lines := strings.Split(code, "\n")

	// Trim leading whitespace/trivia from the start position.
	// AST node positions (Loc.Pos()) may include leading trivia.
	startPos := textRange.Pos()
	endPos := textRange.End()
	for startPos < endPos && startPos < len(code) {
		ch := code[startPos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			startPos++
		} else {
			break
		}
	}

	sl, sc := scanner.GetECMALineAndCharacterOfPosition(sourceFile, startPos)
	el, ec := scanner.GetECMALineAndCharacterOfPosition(sourceFile, endPos)

	// Display range with 1 line of context
	startLine := sl - 1
	if startLine < 0 {
		startLine = 0
	}
	endLine := el + 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Calculate gutter width based on line numbers
	gutterWidth := len(strconv.Itoa(endLine + 1))
	if gutterWidth < 2 {
		gutterWidth = 2
	}

	var sb strings.Builder
	for lineIdx := startLine; lineIdx <= endLine; lineIdx++ {
		lineText := ""
		if lineIdx < len(lines) {
			lineText = lines[lineIdx]
		}
		expandedLine := strings.ReplaceAll(lineText, "\t", "    ")
		fmt.Fprintf(&sb, "  %*d | %s\n", gutterWidth, lineIdx+1, expandedLine)

		if lineIdx < sl || lineIdx > el {
			continue
		}

		// Calculate annotation columns for this line
		aStart := 0
		if lineIdx == sl {
			aStart = sc
		}
		aEnd := len(lineText)
		if lineIdx == el {
			aEnd = ec
		}

		adjustedStart := adjustForTabs(lineText, aStart)
		adjustedEnd := adjustForTabs(lineText, aEnd)

		if adjustedEnd <= adjustedStart {
			continue
		}

		annotationLine := make([]byte, adjustedEnd)
		for i := range annotationLine {
			if i >= adjustedStart {
				annotationLine[i] = marker
			} else {
				annotationLine[i] = ' '
			}
		}

		underline := strings.TrimRight(string(annotationLine), " ")
		if underline != "" {
			fmt.Fprintf(&sb, "  %*s | %s", gutterWidth, "", underline)
			if label != "" {
				sb.WriteString(" ")
				sb.WriteString(label)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// adjustForTabs adjusts a column position to account for tab expansion (tab = 4 spaces).
func adjustForTabs(line string, col int) int {
	adjusted := 0
	for i := 0; i < col && i < len(line); i++ {
		if line[i] == '\t' {
			adjusted += 4
		} else {
			adjusted++
		}
	}
	return adjusted
}

// formatDiagnosticsSnapshot formats diagnostics into a deterministic snapshot string
// with annotated source code showing what is highlighted.
func formatDiagnosticsSnapshot(code string, diagnostics []rule.RuleDiagnostic) string {
	if len(diagnostics) == 0 {
		return "No diagnostics"
	}

	var sb strings.Builder
	for i, d := range diagnostics {
		if i > 0 {
			sb.WriteString("\n")
		}

		// Check if the primary diagnostic range is zeroed/undefined
		hasRange := d.Range.Pos() != d.Range.End() || (d.Range.Pos() != 0 && d.Range.Pos() != -1)

		if hasRange {
			line, column := 0, 0
			endLine, endColumn := 0, 0
			if d.SourceFile != nil {
				// Trim leading trivia from diagnostic start position
				startPos := d.Range.Pos()
				for startPos < d.Range.End() && startPos < len(code) {
					ch := code[startPos]
					if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
						startPos++
					} else {
						break
					}
				}
				lineIdx, colIdx := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, startPos)
				line, column = lineIdx+1, colIdx+1
				// Use inclusive end (last character in range) for display
				inclusiveEnd := d.Range.End()
				if inclusiveEnd > startPos {
					inclusiveEnd--
				}
				endLineIdx, endColIdx := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, inclusiveEnd)
				endLine, endColumn = endLineIdx+1, endColIdx+1
			}

			fmt.Fprintf(&sb, "Diagnostic %d: %s (%d:%d - %d:%d)\n", i+1, d.Message.Id, line, column, endLine, endColumn)
			fmt.Fprintf(&sb, "Message: %s\n", d.Message.Description)

			// Render primary diagnostic range
			annotated := renderSourceAnnotation(code, d.SourceFile, d.Range, '~', "")
			if annotated != "" {
				sb.WriteString(annotated)
			}
		} else {
			fmt.Fprintf(&sb, "Diagnostic %d: %s\n", i+1, d.Message.Id)
			fmt.Fprintf(&sb, "Message: %s\n", d.Message.Description)
		}

		// Render each labeled range as a separate snippet
		for _, lr := range d.LabeledRanges {
			lrLine, lrCol := 0, 0
			lrEndLine, lrEndCol := 0, 0
			if d.SourceFile != nil {
				// Trim leading trivia from labeled range start position
				lrStartPos := lr.Range.Pos()
				for lrStartPos < lr.Range.End() && lrStartPos < len(code) {
					ch := code[lrStartPos]
					if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
						lrStartPos++
					} else {
						break
					}
				}
				lrLineIdx, lrColIdx := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, lrStartPos)
				lrLine, lrCol = lrLineIdx+1, lrColIdx+1
				// Use inclusive end (last character in range) for display
				lrInclusiveEnd := lr.Range.End()
				if lrInclusiveEnd > lrStartPos {
					lrInclusiveEnd--
				}
				lrEndLineIdx, lrEndColIdx := scanner.GetECMALineAndCharacterOfPosition(d.SourceFile, lrInclusiveEnd)
				lrEndLine, lrEndCol = lrEndLineIdx+1, lrEndColIdx+1
			}
			fmt.Fprintf(&sb, "  Label: %s (%d:%d - %d:%d)\n", lr.Label, lrLine, lrCol, lrEndLine, lrEndCol)
			labelAnnotated := renderSourceAnnotation(code, d.SourceFile, lr.Range, '^', lr.Label)
			if labelAnnotated != "" {
				sb.WriteString(labelAnnotated)
			}
		}

		if d.Suggestions != nil && len(*d.Suggestions) > 0 {
			for j, s := range *d.Suggestions {
				fmt.Fprintf(&sb, "  Suggestion %d: [%s] %s\n", j+1, s.Message.Id, s.Message.Description)
			}
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}
