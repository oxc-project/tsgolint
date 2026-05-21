package linter

import (
	"sort"
	"sync"
	"time"
)

type RuleTimingStat struct {
	Duration time.Duration
	Calls    uint64
}

func (s *RuleTimingStat) add(other RuleTimingStat) {
	s.Duration += other.Duration
	s.Calls += other.Calls
}

type RuleTimingRecord struct {
	RuleName string
	Duration time.Duration
	Calls    uint64
}

type RuleTimingStore struct {
	mu      sync.Mutex
	timings map[string]RuleTimingStat
}

func NewRuleTimingStore() *RuleTimingStore {
	return &RuleTimingStore{timings: make(map[string]RuleTimingStat)}
}

func (s *RuleTimingStore) merge(localTimings map[string]RuleTimingStat) {
	if len(localTimings) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for ruleName, stat := range localTimings {
		merged := s.timings[ruleName]
		merged.add(stat)
		s.timings[ruleName] = merged
	}
}

func (s *RuleTimingStore) Collect() []RuleTimingRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := make([]RuleTimingRecord, 0, len(s.timings))
	for ruleName, stat := range s.timings {
		records = append(records, RuleTimingRecord{
			RuleName: ruleName,
			Duration: stat.Duration,
			Calls:    stat.Calls,
		})
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].Duration != records[j].Duration {
			return records[i].Duration > records[j].Duration
		}
		return records[i].RuleName < records[j].RuleName
	})

	return records
}
