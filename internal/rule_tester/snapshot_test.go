package rule_tester

import (
	"slices"
	"testing"
)

func TestSortSnapshotKeys_NaturalNumericOrder(t *testing.T) {
	keys := []string{
		"[TestRule/invalid-21 - 1]",
		"[TestRule/invalid-3 - 1]",
		"[TestRule/invalid-10 - 1]",
		"[TestRule/invalid-2 - 1]",
	}

	sortSnapshotKeys(keys)

	expected := []string{
		"[TestRule/invalid-2 - 1]",
		"[TestRule/invalid-3 - 1]",
		"[TestRule/invalid-10 - 1]",
		"[TestRule/invalid-21 - 1]",
	}

	if !slices.Equal(keys, expected) {
		t.Fatalf("unexpected key order\nexpected: %v\nactual:   %v", expected, keys)
	}
}

func TestSortSnapshotKeys_SortsSnapshotNumberNaturally(t *testing.T) {
	keys := []string{
		"[TestRule/invalid-2 - 10]",
		"[TestRule/invalid-2 - 2]",
		"[TestRule/invalid-2 - 1]",
	}

	sortSnapshotKeys(keys)

	expected := []string{
		"[TestRule/invalid-2 - 1]",
		"[TestRule/invalid-2 - 2]",
		"[TestRule/invalid-2 - 10]",
	}

	if !slices.Equal(keys, expected) {
		t.Fatalf("unexpected key order\nexpected: %v\nactual:   %v", expected, keys)
	}
}
