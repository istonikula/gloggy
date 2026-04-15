package filter

import (
	"testing"
	"time"

	"github.com/istonikula/gloggy/internal/logsource"
)

func mkEntry(level, logger, msg string) logsource.Entry {
	return logsource.Entry{
		IsJSON:  true,
		Time:    time.Now(),
		Level:   level,
		Logger:  logger,
		Msg:     msg,
		Raw:     []byte(`{"level":"` + level + `"}`),
	}
}

var testEntries = []logsource.Entry{
	mkEntry("ERROR", "app", "disk full"),
	mkEntry("WARN", "app", "low memory"),
	mkEntry("INFO", "noisy", "heartbeat"),
	mkEntry("ERROR", "noisy", "heartbeat"),
	mkEntry("DEBUG", "core", "trace"),
	mkEntry("WARN", "noisy", "retry"),
}

func eqInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// T-019: R3.1 — one include filter level=ERROR shows only ERROR entries
func TestApply_SingleInclude_LevelError(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: true})
	got := Apply(fs, testEntries)
	want := []int{0, 3}
	if !eqInts(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// T-019: R3.2 — two include filters show entries matching either
func TestApply_TwoIncludes(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: true})
	fs.Add(Filter{Field: "level", Pattern: "WARN", Mode: Include, Enabled: true})
	got := Apply(fs, testEntries)
	want := []int{0, 1, 3, 5}
	if !eqInts(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// T-019: R3.3 — exclude filter hides matching, all others pass
func TestApply_ExcludeOnly(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "logger", Pattern: "noisy", Mode: Exclude, Enabled: true})
	got := Apply(fs, testEntries)
	want := []int{0, 1, 4}
	if !eqInts(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// T-019: R3.4 — include + exclude combined
func TestApply_IncludeAndExclude(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: true})
	fs.Add(Filter{Field: "msg", Pattern: "heartbeat", Mode: Exclude, Enabled: true})
	got := Apply(fs, testEntries)
	// index 0 (ERROR, app, disk full) passes; index 3 (ERROR, noisy, heartbeat) excluded
	want := []int{0}
	if !eqInts(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// T-019: R3.5 — disabled filters have no effect
func TestApply_DisabledFiltersIgnored(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: false})
	fs.Add(Filter{Field: "logger", Pattern: "noisy", Mode: Exclude, Enabled: false})
	got := Apply(fs, testEntries)
	want := []int{0, 1, 2, 3, 4, 5}
	if !eqInts(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// T-020: R7.1 — index contains exactly passing entries
func TestFilteredIndex_ExactlyPassingEntries(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "WARN", Mode: Include, Enabled: true})
	fi := NewFilteredIndex(fs, testEntries)
	want := []int{1, 5}
	if !eqInts(fi.Indices, want) {
		t.Fatalf("got %v, want %v", fi.Indices, want)
	}
}

// T-020: R7.2 — index preserves original order
func TestFilteredIndex_PreservesOrder(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: true})
	fs.Add(Filter{Field: "level", Pattern: "WARN", Mode: Include, Enabled: true})
	fi := NewFilteredIndex(fs, testEntries)
	for i := 1; i < len(fi.Indices); i++ {
		if fi.Indices[i] <= fi.Indices[i-1] {
			t.Fatalf("indices not in original order: %v", fi.Indices)
		}
	}
}

// T-020: R7.3 — Recompute updates on filter change
func TestFilteredIndex_Recompute(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "level", Pattern: "ERROR", Mode: Include, Enabled: true})
	fi := NewFilteredIndex(fs, testEntries)
	if !eqInts(fi.Indices, []int{0, 3}) {
		t.Fatalf("initial: got %v, want [0,3]", fi.Indices)
	}

	fs.Remove(id)
	fs.Add(Filter{Field: "level", Pattern: "DEBUG", Mode: Include, Enabled: true})
	fi.Recompute(fs, testEntries)
	if !eqInts(fi.Indices, []int{4}) {
		t.Fatalf("after recompute: got %v, want [4]", fi.Indices)
	}
}
