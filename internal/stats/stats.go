package stats

import (
	"time"
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
	TsgolintVersion string
	TsgoVersion     string
	ThreadCount     int
	TsconfigCount   int
	Programs        []Program
	Rules           map[string]time.Duration
	Compile         time.Duration
	LintWall        time.Duration
	LintCPU         time.Duration
	Total           time.Duration
}

func NewReport(tsgolintVersion, tsgoVersion string) *Report {
	return &Report{
		TsgolintVersion: tsgolintVersion,
		TsgoVersion:     tsgoVersion,
		Programs:        make([]Program, 0),
		Rules:           make(map[string]time.Duration),
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
