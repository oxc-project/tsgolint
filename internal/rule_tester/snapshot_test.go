package rule_tester

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSnapshotBatchWritesOnceAfterSubtests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "batched.snap")
	t.Run("batch", func(t *testing.T) {
		snapshotRegistry.beginBatch(t, path)
		for _, key := range []string{"[test/invalid-2 - 1]", "[test/invalid-1 - 1]"} {
			t.Run(key, func(t *testing.T) {
				snapshotRegistry.matchSnapshot(t, path, key, "diagnostic", false)
			})
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("snapshot was written before parent cleanup: %v", err)
		}
	})
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := parseSnapshotFile(string(data))
	if len(entries) != 2 || entries["[test/invalid-1 - 1]"] != "diagnostic" || entries["[test/invalid-2 - 1]"] != "diagnostic" {
		t.Fatalf("unexpected batched snapshot entries: %#v", entries)
	}
}

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
