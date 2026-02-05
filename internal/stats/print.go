package stats

import (
	"fmt"
	"io"
	"path/filepath"
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

// PrintReport prints the stats report to w, using currentDir to display relative paths.
func PrintReport(w io.Writer, s *Report, cwd string) {
	fmt.Fprint(w, "\n")

	fmt.Fprintln(w, "Version:")
	printVersionRow(w, "tsgolint", s.TsgolintVersion)
	printVersionRow(w, "tsgo", s.TsgoVersion)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Typecheck:")
	printTypecheckSection(w, s, cwd)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Lint:")
	printLintSection(w, s)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Summary:")
	printSummarySection(w, s)
	fmt.Fprintln(w)
}

func displayName(cwd, configPath string) string {
	if cwd == "" {
		return configPath
	}
	rel, err := filepath.Rel(cwd, configPath)
	if err != nil {
		return configPath
	}
	return rel
}

func printVersionRow(w io.Writer, name, version string) {
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, name, colVersion, version)
}

func printTypecheckSection(w io.Writer, s *Report, cwd string) {
	programs := make([]Program, len(s.Programs))
	copy(programs, s.Programs)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Time > programs[j].Time
	})

	fmt.Fprintf(w, "%s%-*s%*s%*s\n", colIndent, colName, "Program", colValue, "Wall Time", colFiles, "Files")

	totalFiles := 0
	for _, p := range programs {
		name := displayName(cwd, p.Name)
		if len(name) > colName {
			name = "..." + name[len(name)-colName+3:]
		}
		fmt.Fprintf(w, "%s%-*s%*s%*d\n", colIndent, colName, name, colValue, formatDuration(p.Time), colFiles, p.FileCount)
		totalFiles += p.FileCount
	}

	fmt.Fprintf(w, "%s%s\n", colIndent, strings.Repeat("─", colName+colValue+colFiles))
	fmt.Fprintf(w, "%s%-*s%*s%*d\n", colIndent, colName, "Total", colValue, formatDuration(s.Compile), colFiles, totalFiles)
}

func printLintSection(w io.Writer, s *Report) {
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

func printSummarySection(w io.Writer, s *Report) {
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Category", colValue, "Wall time")
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "typecheck", colValue, formatDuration(s.Compile))
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "lint", colValue, formatDuration(s.LintWall))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
