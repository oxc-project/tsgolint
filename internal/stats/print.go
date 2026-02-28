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
	colIndent = "    "
	colName   = 56 // first column width (name/label), positions 5-60 on screen
	colValue  = 10 // second column width (time/value)
	colFiles  = 10 // third column width (files count)

	maxRules = 5
)

func PrintReport(w io.Writer, s *Report, cwd string) {
	if s == nil {
		return
	}
	fmt.Fprint(w, "\n")

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

func printTypecheckSection(w io.Writer, s *Report, cwd string) {
	programs := make([]Program, len(s.Programs))
	copy(programs, s.Programs)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Time > programs[j].Time
	})

	names := make([]string, len(programs))
	for i, p := range programs {
		name := displayName(cwd, p.Name)
		if len(name) > colName {
			name = "..." + name[len(name)-colName+3:]
		}
		names[i] = name
	}

	fmt.Fprintf(w, "%s%-*s%*s%*s\n", colIndent, colName, "Program", colValue, "Wall Time", colFiles, "Files")
	for i, p := range programs {
		fmt.Fprintf(w, "%s%-*s%*s%*d\n", colIndent, colName, names[i], colValue, formatDuration(p.Time), colFiles, p.FileCount)
	}
	fmt.Fprintf(w, "%s%s\n", colIndent, strings.Repeat("─", colName+colValue+colFiles))
	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Total", colValue, formatDuration(s.Compile))
}

func printLintSection(w io.Writer, s *Report) {
	rules := make([]Rule, 0, len(s.Rules))
	for name, duration := range s.Rules {
		rules = append(rules, Rule{Name: name, Time: duration})
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Time > rules[j].Time
	})

	displayCount := min(maxRules, len(rules))
	hiddenCount := len(rules) - displayCount

	var totalRules time.Duration
	for _, r := range rules {
		totalRules += r.Time
	}

	var hiddenTime time.Duration
	for i := displayCount; i < len(rules); i++ {
		hiddenTime += rules[i].Time
	}

	traversal := max(s.LintCPU-totalRules, 0)

	fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, "Rule", colValue, "CPU Time")
	for i := range displayCount {
		fmt.Fprintf(w, "%s%-*s%*s\n", colIndent, colName, rules[i].Name, colValue, formatDuration(rules[i].Time))
	}
	if hiddenCount > 0 {
		fmt.Fprintf(w, "%s... %d more rules (%s)\n", colIndent, hiddenCount, formatDuration(hiddenTime))
	}
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
