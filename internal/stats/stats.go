package stats

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// ProgramStat holds timing and file count for a single TypeScript program
type ProgramStat struct {
	Name      string
	Time      time.Duration
	FileCount int
}

// RuleStat holds timing for a single lint rule
type RuleStat struct {
	Name string
	Time time.Duration
}

// LintStats collects performance statistics for a lint run
type LintStats struct {
	TsgolintVersion string
	TsgoVersion     string
	ThreadCount     int
	TsconfigCount   int
	Programs        []ProgramStat
	Rules           map[string]time.Duration
	CompileTime     time.Duration
	LintTime        time.Duration
}

// NewLintStats creates a new LintStats instance
func NewLintStats(tsgolintVersion, tsgoVersion string, threadCount int) *LintStats {
	return &LintStats{
		TsgolintVersion: tsgolintVersion,
		TsgoVersion:     tsgoVersion,
		ThreadCount:     threadCount,
		Programs:        make([]ProgramStat, 0),
		Rules:           make(map[string]time.Duration),
	}
}

// AddProgramStat adds timing stats for a program
func (s *LintStats) AddProgramStat(name string, duration time.Duration, fileCount int) {
	s.Programs = append(s.Programs, ProgramStat{
		Name:      name,
		Time:      duration,
		FileCount: fileCount,
	})
	s.TsconfigCount++
	s.CompileTime += duration
}

// AddRuleTime adds execution time for a rule (thread-safe accumulation should be done externally)
func (s *LintStats) AddRuleTime(ruleName string, duration time.Duration) {
	s.Rules[ruleName] += duration
	s.LintTime += duration
}

// Enabled returns true if OXC_TSGOLINT_STATS environment variable is set
func Enabled() bool {
	return os.Getenv("OXC_TSGOLINT_STATS") != ""
}

// Print outputs the stats to the given writer
func (s *LintStats) Print(w io.Writer) {
	// Header
	fmt.Fprintf(w, "\ntsgolint stats (%d tsconfigs, %d threads)\n\n", s.TsconfigCount, s.ThreadCount)

	// Version section
	fmt.Fprintln(w, "Version:")
	s.printVersionRow(w, "tsgolint", s.TsgolintVersion)
	s.printVersionRow(w, "tsgo", s.TsgoVersion)
	fmt.Fprintln(w)

	// Typecheck section
	fmt.Fprintln(w, "Typecheck:")
	s.printTypecheckSection(w)
	fmt.Fprintln(w)

	// Lint section
	fmt.Fprintln(w, "Lint:")
	s.printLintSection(w)
	fmt.Fprintln(w)

	// Summary section
	fmt.Fprintln(w, "Summary:")
	s.printSummaryRow(w, "Compile", s.CompileTime)
	s.printSummaryRow(w, "Lint", s.LintTime)
}

func (s *LintStats) printVersionRow(w io.Writer, name, version string) {
	fmt.Fprintf(w, "    %-36s %s\n", name+":", version)
}

func (s *LintStats) printTypecheckSection(w io.Writer) {
	// Sort programs by time (descending)
	programs := make([]ProgramStat, len(s.Programs))
	copy(programs, s.Programs)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Time > programs[j].Time
	})

	// Calculate column widths
	maxNameLen := len("Program")
	for _, p := range programs {
		if len(p.Name) > maxNameLen {
			maxNameLen = len(p.Name)
		}
	}
	if maxNameLen > 40 {
		maxNameLen = 40
	}

	// Header
	fmt.Fprintf(w, "    %-*s  %10s  %8s\n", maxNameLen, "Program", "Time", "Files")

	// Programs
	totalFiles := 0
	for _, p := range programs {
		name := p.Name
		if len(name) > maxNameLen {
			name = "..." + name[len(name)-maxNameLen+3:]
		}
		fmt.Fprintf(w, "    %-*s  %10s  %8d\n", maxNameLen, name, formatDuration(p.Time), p.FileCount)
		totalFiles += p.FileCount
	}

	// Separator and total
	fmt.Fprintf(w, "    %s\n", strings.Repeat("-", maxNameLen+22))
	fmt.Fprintf(w, "    %-*s  %10s  %8d\n", maxNameLen, "Total", formatDuration(s.CompileTime), totalFiles)
}

func (s *LintStats) printLintSection(w io.Writer) {
	// Convert map to slice and sort by time (descending)
	rules := make([]RuleStat, 0, len(s.Rules))
	for name, duration := range s.Rules {
		rules = append(rules, RuleStat{Name: name, Time: duration})
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Time > rules[j].Time
	})

	// Calculate column width
	maxNameLen := len("Rule")
	for _, r := range rules {
		if len(r.Name) > maxNameLen {
			maxNameLen = len(r.Name)
		}
	}
	if maxNameLen > 40 {
		maxNameLen = 40
	}

	// Header
	fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, "Rule", "Time")

	// Show top 5 rules
	displayCount := min(5, len(rules))
	for i := 0; i < displayCount; i++ {
		fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, rules[i].Name, formatDuration(rules[i].Time))
	}

	// Collapsed remaining rules
	if len(rules) > 5 {
		remainingCount := len(rules) - 5
		var remainingTime time.Duration
		for i := 5; i < len(rules); i++ {
			remainingTime += rules[i].Time
		}
		fmt.Fprintf(w, "    ... %d more rules (%s)\n", remainingCount, formatDuration(remainingTime))
	}

	// Separator and total
	fmt.Fprintf(w, "    %s\n", strings.Repeat("-", maxNameLen+12))
	fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, "Total", formatDuration(s.LintTime))
}

func (s *LintStats) printSummaryRow(w io.Writer, name string, duration time.Duration) {
	fmt.Fprintf(w, "    %-36s %s\n", name+":", formatDuration(duration))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
