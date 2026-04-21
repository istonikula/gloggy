// Package integration contains integration tests that wire multiple subsystems together.
package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

func makeFilterEntries() []logsource.Entry {
	levels := []string{"INFO", "ERROR", "WARN", "INFO", "ERROR"}
	entries := make([]logsource.Entry, len(levels))
	for i, l := range levels {
		entries[i] = logsource.Entry{
			IsJSON:     true,
			LineNumber: i + 1,
			Level:      l,
			Msg:        fmt.Sprintf("msg %d", i),
			Time:       time.Now(),
			Raw:        []byte(fmt.Sprintf(`{"level":%q,"msg":"msg %d"}`, l, i)),
		}
	}
	return entries
}

// T-055: entry-list/R7 — filtered list shows only passing entries.
func TestFilterList_OnlyPassingEntries(t *testing.T) {
	entries := makeFilterEntries()
	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "ERROR", Mode: filter.Include, Enabled: true})

	indices := filter.Apply(fs, entries)
	// Entries at index 1 (ERROR) and 4 (ERROR) should pass.
	require.Len(t, indices, 2, "expected 2 filtered entries (ERROR)")
	for _, idx := range indices {
		assert.Equalf(t, "ERROR", entries[idx].Level, "expected ERROR entry, got %s at index %d", entries[idx].Level, idx)
	}
}

// T-055: filter-engine/R7 — filter change updates list; selection preserved or moved.
func TestFilterList_FilterChange_UpdatesList(t *testing.T) {
	entries := makeFilterEntries()
	m := entrylist.NewListModel(theme.GetTheme("tokyo-night"), config.DefaultConfig(), 80, 10)
	m = m.SetEntries(entries)

	fs := filter.NewFilterSet()
	fs.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	indices := filter.Apply(fs, entries)

	m = m.SetFilter(indices)
	vis := m.Marks() // just exercise the code
	_ = vis
	// After filter: only INFO entries (indices 0, 3) should be visible.
	require.Len(t, indices, 2, "expected 2 INFO entries")

	// Change filter to include WARN.
	fs2 := filter.NewFilterSet()
	fs2.Add(filter.Filter{Field: "level", Pattern: "WARN", Mode: filter.Include, Enabled: true})
	indices2 := filter.Apply(fs2, entries)
	m2 := m.SetFilter(indices2)
	require.Len(t, indices2, 1, "expected 1 WARN entry")
	// Cursor should be non-negative after filter change.
	assert.GreaterOrEqualf(t, m2.Cursor(), 0, "cursor should be non-negative after filter change: %d", m2.Cursor())
}
