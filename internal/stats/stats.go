package stats

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	colIndent  = "    "
	colName    = 56 // first column width (name/label), positions 5-60 on screen
	colValue   = 10 // second column width (time/value)
	colFiles   = 10 // third column width (files count)
	colVersion = colValue + colFiles
)

type Program struct {
	Name      string
	Time      time.Duration
	FileCount int
}

type Rule struct {
	Name string
	Time time.Duration
}

type Report struct {
	TsgolintVersion  string
	TsgoVersion      string
	ThreadCount      int
	TsconfigCount    int
	CurrentDirectory string
	Programs         []Program
	Rules            map[string]time.Duration
	Compile          time.Duration
	LintWall         time.Duration
	LintCPU          time.Duration
	Total            time.Duration
}

func NewReport(tsgolintVersion, tsgoVersion string, currentDirectory string) *Report {
	return &Report{
		TsgolintVersion:  tsgolintVersion,
		TsgoVersion:      tsgoVersion,
		CurrentDirectory: currentDirectory,
		Programs:         make([]Program, 0),
		Rules:            make(map[string]time.Duration),
	}
}

func (s *Report) AddProgram(name string, d time.Duration, fileCount int) {
	s.Programs = append(s.Programs, Program{
		Name:      name,
		Time:      d,
		FileCount: fileCount,
	})
	s.TsconfigCount++
	s.Compile += d
}

func (s *Report) AddRule(rule string, d time.Duration) {
	s.Rules[rule] += d
}

func (s *Report) AddLintWall(d time.Duration) {
	s.LintWall += d
}

func (s *Report) AddLintCPU(d time.Duration) {
	s.LintCPU += d
}

func (s *Report) SetTotal(d time.Duration) {
	s.Total = d
}

func Enabled() bool {
	return os.Getenv("OXC_TSGOLINT_STATS") != ""
}

func (s *Report) displayName(name string) string {
	if s.CurrentDirectory == "" {
		return name
	}
	prefix := s.CurrentDirectory
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if rel, ok := strings.CutPrefix(name, prefix); ok {
		return rel
	}
	return name
}

func (s *Report) Print(w io.Writer) {
	fmt.Fprint(w, "\n")

	fmt.Fprintln(w, "Version:")
	s.printVersionRow(w, "tsgolint", s.TsgolintVersion)
	s.printVersionRow(w, "tsgo", s.TsgoVersion)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Typecheck:")
	s.printTypecheckSection(w)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Lint:")
	s.printLintSection(w)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Summary:")
	s.printSummarySection(w)
	fmt.Fprintln(w)
}

func (s *Report) printVersionRow(w io.Writer, name, version string) {
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, name, colVersion, version)
}

func (s *Report) printTypecheckSection(w io.Writer) {
	programs := make([]Program, len(s.Programs))
	copy(programs, s.Programs)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Time > programs[j].Time
	})

	fmt.Fprintf(w, "%s%-*s%*s%*s\n", colIndent, colName, "Program", colValue, "Wall Time", colFiles, "Files")

	totalFiles := 0
	for _, p := range programs {
		name := s.displayName(p.Name)
		if len(name) > colName {
			name = "..." + name[len(name)-colName+3:]
		}
		fmt.Fprintf(w, "%s%-*s%*s%*d\n", colIndent, colName, name, colValue, formatDuration(p.Time), colFiles, p.FileCount)
		totalFiles += p.FileCount
	}

	fmt.Fprintf(w, "%s%s\n", colIndent, strings.Repeat("─", colName+colValue+colFiles))
	fmt.Fprintf(w, "%s%-*s%*s%*d\n", colIndent, colName, "Total", colValue, formatDuration(s.Compile), colFiles, totalFiles)
}

func (s *Report) printLintSection(w io.Writer) {
	rules := make([]Rule, 0, len(s.Rules))
	for name, duration := range s.Rules {
		rules = append(rules, Rule{Name: name, Time: duration})
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Time > rules[j].Time
	})

	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Rule", colValue, "CPU Time")

	displayCount := min(5, len(rules))
	for i := range displayCount {
		fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, rules[i].Name, colValue, formatDuration(rules[i].Time))
	}

	if len(rules) > 5 {
		remainingCount := len(rules) - 5
		var remainingTime time.Duration
		for i := 5; i < len(rules); i++ {
			remainingTime += rules[i].Time
		}
		fmt.Fprintf(w, "%s... %d more rules (%s)\n", colIndent, remainingCount, formatDuration(remainingTime))
	}

	var totalRules time.Duration
	for _, r := range rules {
		totalRules += r.Time
	}
	traversal := max(s.LintCPU-totalRules, 0)

	fmt.Fprintf(w, "%s%s\n", colIndent, strings.Repeat("─", colName+colValue))
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Traversal+overhead", colValue, formatDuration(traversal))
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Total", colValue, formatDuration(s.LintCPU))
}

func (s *Report) printSummarySection(w io.Writer) {
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Category", colValue, "Wall time")
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "typecheck", colValue, formatDuration(s.Compile))
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "lint", colValue, formatDuration(s.LintWall))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
