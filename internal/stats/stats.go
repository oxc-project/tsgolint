package stats

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type ProgramStat struct {
	Name      string
	Time      time.Duration
	FileCount int
}

type RuleStat struct {
	Name string
	Time time.Duration
}

type LintStats struct {
	TsgolintVersion string
	TsgoVersion     string
	ThreadCount     int
	TsconfigCount   int
	Programs        []ProgramStat
	Rules           map[string]time.Duration
	CompileTime     time.Duration
	LintTime        time.Duration
	LintCPUTime     time.Duration
	TotalWallTime   time.Duration
}

func NewLintStats(tsgolintVersion, tsgoVersion string, threadCount int) *LintStats {
	return &LintStats{
		TsgolintVersion: tsgolintVersion,
		TsgoVersion:     tsgoVersion,
		ThreadCount:     threadCount,
		Programs:        make([]ProgramStat, 0),
		Rules:           make(map[string]time.Duration),
	}
}

func (s *LintStats) AddProgramStat(name string, duration time.Duration, fileCount int) {
	s.Programs = append(s.Programs, ProgramStat{
		Name:      name,
		Time:      duration,
		FileCount: fileCount,
	})
	s.TsconfigCount++
	s.CompileTime += duration
}

func (s *LintStats) AddRuleTime(ruleName string, duration time.Duration) {
	s.Rules[ruleName] += duration
}

func (s *LintStats) AddLintTime(duration time.Duration) {
	s.LintTime += duration
}

func (s *LintStats) AddLintCPUTime(duration time.Duration) {
	s.LintCPUTime += duration
}

func (s *LintStats) SetTotalTime(duration time.Duration) {
	s.TotalWallTime = duration
}

func Enabled() bool {
	return os.Getenv("OXC_TSGOLINT_STATS") != ""
}

func (s *LintStats) Print(w io.Writer) {
	fmt.Fprintf(w, "\ntsgolint stats (%d tsconfigs, %d threads)\n\n", s.TsconfigCount, s.ThreadCount)

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
	s.printSummaryRow(w, "Compile", s.CompileTime)
	s.printSummaryRow(w, "Lint", s.LintTime)
	other := s.TotalWallTime - s.CompileTime - s.LintTime
	if other < 0 {
		other = 0
	}
	s.printSummaryRow(w, "Other", other)
	s.printSummaryRow(w, "Total", s.TotalWallTime)
}

func (s *LintStats) printVersionRow(w io.Writer, name, version string) {
	fmt.Fprintf(w, "    %-36s %s\n", name+":", version)
}

func (s *LintStats) printTypecheckSection(w io.Writer) {
	programs := make([]ProgramStat, len(s.Programs))
	copy(programs, s.Programs)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].Time > programs[j].Time
	})

	maxNameLen := len("Program")
	for _, p := range programs {
		if len(p.Name) > maxNameLen {
			maxNameLen = len(p.Name)
		}
	}
	if maxNameLen > 40 {
		maxNameLen = 40
	}

	fmt.Fprintf(w, "    %-*s  %10s  %8s\n", maxNameLen, "Program", "Time", "Files")

	totalFiles := 0
	for _, p := range programs {
		name := p.Name
		if len(name) > maxNameLen {
			name = "..." + name[len(name)-maxNameLen+3:]
		}
		fmt.Fprintf(w, "    %-*s  %10s  %8d\n", maxNameLen, name, formatDuration(p.Time), p.FileCount)
		totalFiles += p.FileCount
	}

	fmt.Fprintf(w, "    %s\n", strings.Repeat("-", maxNameLen+22))
	fmt.Fprintf(w, "    %-*s  %10s  %8d\n", maxNameLen, "Total", formatDuration(s.CompileTime), totalFiles)
}

func (s *LintStats) printLintSection(w io.Writer) {
	rules := make([]RuleStat, 0, len(s.Rules))
	for name, duration := range s.Rules {
		rules = append(rules, RuleStat{Name: name, Time: duration})
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Time > rules[j].Time
	})

	maxNameLen := len("Rule")
	for _, r := range rules {
		if len(r.Name) > maxNameLen {
			maxNameLen = len(r.Name)
		}
	}
	if maxNameLen > 40 {
		maxNameLen = 40
	}

	fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, "Rule", "Time")

	displayCount := min(5, len(rules))
	for i := range displayCount {
		fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, rules[i].Name, formatDuration(rules[i].Time))
	}

	if len(rules) > 5 {
		remainingCount := len(rules) - 5
		var remainingTime time.Duration
		for i := 5; i < len(rules); i++ {
			remainingTime += rules[i].Time
		}
		fmt.Fprintf(w, "    ... %d more rules (%s)\n", remainingCount, formatDuration(remainingTime))
	}

	var totalRules time.Duration
	for _, r := range rules {
		totalRules += r.Time
	}
	traversal := s.LintCPUTime - totalRules
	if traversal < 0 {
		traversal = 0
	}

	fmt.Fprintf(w, "    %s\n", strings.Repeat("-", maxNameLen+12))
	fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, "Traversal+overhead", formatDuration(traversal))
	fmt.Fprintf(w, "    %-*s  %10s\n", maxNameLen, "Total", formatDuration(s.LintCPUTime))
}

func (s *LintStats) printSummaryRow(w io.Writer, name string, duration time.Duration) {
	fmt.Fprintf(w, "    %-36s %s\n", name+":", formatDuration(duration))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
