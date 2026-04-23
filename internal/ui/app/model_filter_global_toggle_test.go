package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// ---------- V32 (a): F is gated on pane-search input mode (V14) ----------

// TestModel_F_DoesNotToggle_DuringListSearchInput verifies V14/V32(a):
// while the list search is in input mode, `F` extends the query instead of
// triggering the global filter toggle — same policy as `?`/`T`/`t`.
func TestModel_F_DoesNotToggle_DuringListSearchInput(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	require.False(t, m.filterSet.IsGloballyDisabled(), "precondition: filters not globally-disabled")

	m = key(m, "/")
	m = key(m, "a")
	m = key(m, "b")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search active")
	require.True(t, m.list.Search().InputMode(), "precondition: list search in input mode")

	m = key(m, "F")

	assert.Falsef(t, m.filterSet.IsGloballyDisabled(),
		"F must NOT toggle filters while list search is in input mode (V14/V32a)")
	assert.Equalf(t, "abF", m.list.Search().Query(),
		"F should extend the query to %q, got %q", "abF", m.list.Search().Query())
}

// TestModel_F_DoesNotToggle_DuringPaneSearchInput verifies V14/V32(a) for the
// detail pane: with pane search in input mode, `F` is consumed as a query
// char and filters are unchanged.
func TestModel_F_DoesNotToggle_DuringPaneSearchInput(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m = m.openPane(entries[0])
	m = key(m, "tab")
	m = key(m, "/")
	require.True(t, m.paneSearch.IsActive(), "precondition: pane search active")
	require.Equalf(t, detailpane.SearchModeInput, m.paneSearch.Mode(),
		"precondition: pane search in input mode")

	m = key(m, "F")

	assert.Falsef(t, m.filterSet.IsGloballyDisabled(),
		"F must NOT toggle filters while pane search is in input mode (V14/V32a)")
	assert.Equalf(t, "F", m.paneSearch.Query(),
		"F should extend the pane-search query, got %q", m.paneSearch.Query())
}

// ---------- V32 (b): F from clean state disables all filters ----------

// TestModel_F_DisablesAllFilters_AndNotice verifies V32(b): first `F` press
// saves per-filter Enabled + disables all; notice is "filters disabled"; the
// filter index recomputes so the filtered visible-count changes.
func TestModel_F_DisablesAllFilters_AndNotice(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(10))
	// Two filters, both enabled. Single include filter that matches nothing
	// hides every entry; add a second enabled filter to prove ToggleAll
	// disables them all (not just the first).
	m.filterSet.Add(filter.Filter{Field: "msg", Pattern: "never-matches-xyz", Mode: filter.Include, Enabled: true})
	m.filterSet.Add(filter.Filter{Field: "logger", Pattern: "also-never", Mode: filter.Include, Enabled: true})
	m = m.refilter()
	require.Equalf(t, 0, m.cachedVisibleCount,
		"precondition: include-only filters with non-matching patterns hide all entries")

	m = key(m, "F")

	require.Truef(t, m.filterSet.IsGloballyDisabled(),
		"F should mark FilterSet globally-disabled")
	for _, f := range m.filterSet.GetAll() {
		assert.Falsef(t, f.Enabled, "filter %q should be disabled after F", f.Field)
	}
	assert.Equalf(t, 10, m.cachedVisibleCount,
		"refilter should restore the full entry set once every filter is disabled")
	assert.Truef(t, m.keyhints.HasNotice(),
		"F should surface a V15-pattern notice; none present")
	assert.Containsf(t, m.view(), filterToggleDisabledNotice,
		"rendered view should contain %q", filterToggleDisabledNotice)
}

// ---------- V32 (c): second F restores saved per-filter state ----------

// TestModel_F_SecondPress_RestoresSavedState verifies V32(c): after disabling
// globally, a second `F` re-enables filters that were previously Enabled and
// keeps filters that were previously Disabled still disabled.
func TestModel_F_SecondPress_RestoresSavedState(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(10))
	// One enabled, one individually disabled pre-toggle.
	m.filterSet.Add(filter.Filter{Field: "msg", Pattern: "off", Mode: filter.Include, Enabled: false})
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m = m.refilter()

	m = key(m, "F")
	require.Truef(t, m.filterSet.IsGloballyDisabled(), "precondition: first F disabled all")

	m = key(m, "F")

	assert.Falsef(t, m.filterSet.IsGloballyDisabled(),
		"second F should clear globally-disabled flag")
	filters := m.filterSet.GetAll()
	require.Lenf(t, filters, 2, "expected 2 filters; got %d", len(filters))
	assert.Falsef(t, filters[0].Enabled,
		"pre-toggle-disabled filter must stay disabled after restore")
	assert.Truef(t, filters[1].Enabled,
		"pre-toggle-enabled filter must be restored")
	assert.Containsf(t, m.view(), filterToggleRestoredNotice,
		"second F should surface %q notice", filterToggleRestoredNotice)
}

// ---------- V32 (d): 0 filters = notice-only, no state change ----------

// TestModel_F_NoFilters_NoticeOnly verifies V32(d): with an empty FilterSet,
// `F` emits the "no filters" notice, changes no state, and does not flip the
// globally-disabled flag.
func TestModel_F_NoFilters_NoticeOnly(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	require.Emptyf(t, m.filterSet.GetAll(), "precondition: no filters")

	m = key(m, "F")

	assert.Falsef(t, m.filterSet.IsGloballyDisabled(),
		"F on empty FilterSet must NOT flip globally-disabled")
	assert.Emptyf(t, m.filterSet.GetAll(), "FilterSet must stay empty")
	assert.Containsf(t, m.view(), filterToggleNoFiltersNotice,
		"empty-FilterSet F should surface %q notice", filterToggleNoFiltersNotice)
}

// view renders the model and returns it as a string, for notice assertions.
func (m Model) view() string {
	var sb strings.Builder
	sb.WriteString(m.View())
	return sb.String()
}
