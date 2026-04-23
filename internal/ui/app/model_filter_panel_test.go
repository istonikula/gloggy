package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// ---------- T30 / V33 VIEW-AXIS: f renders filter panel overlay ----------

// Empty-state: pressing `f` with zero filters renders the discovery-hint
// empty copy inside a full-screen overlay. Guards B10: previously `f`
// flipped focus but Model.View() never composed the panel ∴ live TUI
// showed an unchanged frame.
func TestModel_F_RendersEmptyPanelOverlay_V33(t *testing.T) {
	m := resize(newModel(), 90, 30)
	m = m.SetEntries([]logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	require.Empty(t, m.filterSet.GetAll(), "precondition: no filters")

	m = key(m, "f")

	require.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"precondition: f transfers focus to filter panel")

	view := ansi.Strip(m.View())
	assert.Containsf(t, view, "Filters",
		"overlay should show a title when panel focused")
	assert.Containsf(t, view, "no filters",
		"empty-state copy should mention zero filters")
	assert.Containsf(t, view, "click a field",
		"empty-state should surface the add-filter discovery hint")
	assert.Containsf(t, view, "Esc",
		"overlay footer should surface Esc-to-close keyhint")
}

// Non-empty: seed 2 filters, press `f`, assert both rows render with
// their field:pattern + enabled/disabled glyph.
func TestModel_F_RendersFilterRows_V33(t *testing.T) {
	m := resize(newModel(), 90, 30)
	m = m.SetEntries([]logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m.filterSet.Add(filter.Filter{Field: "msg", Pattern: "boom", Mode: filter.Exclude, Enabled: false})

	m = key(m, "f")

	view := ansi.Strip(m.View())
	assert.Containsf(t, view, "level:INFO", "first filter row")
	assert.Containsf(t, view, "msg:boom", "second filter row")
	assert.Containsf(t, view, "[x]", "enabled-indicator glyph")
	assert.Containsf(t, view, "[ ]", "disabled-indicator glyph")
	assert.Containsf(t, view, "include", "include mode label")
	assert.Containsf(t, view, "exclude", "exclude mode label")
}

// Esc closes the overlay: after open+Esc, m.View() should not contain
// the overlay's title + keyhints footer; focus returns to entry list.
func TestModel_F_Esc_RestoresPreOpenFrame_V33(t *testing.T) {
	m := resize(newModel(), 90, 30)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "f")
	require.Equal(t, appshell.FocusFilterPanel, m.focus, "precondition: panel open")

	m = key(m, "esc")

	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"Esc should return focus to entry list")
	view := ansi.Strip(m.View())
	assert.NotContainsf(t, view, "Filters\n",
		"overlay title must not appear after Esc; got %q", firstLines(view, 3))
	assert.NotContainsf(t, view, "Navigate filters",
		"overlay keyhints footer must not appear after Esc")
}

// V28: overlay View MUST NOT emit any `\t` — bubbletea's diff-renderer
// leaves tab-skipped cells populated with bytes from the previous frame,
// bleeding list rows into the overlay. Guards the same class as B6.
func TestModel_F_OverlayView_NoTabs_V28(t *testing.T) {
	m := resize(newModel(), 90, 30)
	m = m.SetEntries(makeEntries(1))
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})

	m = key(m, "f")

	view := m.View()
	assert.NotContainsf(t, view, "\t",
		"filter-panel overlay View() must pad with spaces, never `\\t` (V28)")
}

// V14 / V33: pane-search in input mode MUST consume `f` as a query
// char — not transfer focus to the filter panel. `f` is already
// focus-scoped (only the FocusEntryList branch opens the panel), so
// while FocusDetailPane + pane-search input mode, the key routes into
// paneSearch.Update. Test codifies the behavior so a future refactor
// that lifts `f` to a global doesn't regress V14.
func TestModel_F_DuringPaneSearchInput_DoesNotOpenPanel_V14(t *testing.T) {
	m := resize(newModel(), 90, 30)
	entries := []logsource.Entry{jsonEntryLvl(1, "INFO", "filter-me")}
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = setFocus(m, appshell.FocusDetailPane)
	m = key(m, "/")
	require.Truef(t, m.paneSearch.IsActive(),
		"precondition: pane search active")
	require.Equalf(t, detailpane.SearchModeInput, m.paneSearch.Mode(),
		"precondition: pane search in input mode")

	m = key(m, "f")

	assert.NotEqualf(t, appshell.FocusFilterPanel, m.focus,
		"V14: f during pane-search input mode must not open filter panel")
}

// firstLines returns the first n newline-separated lines of s for
// assertion failure diagnostics.
func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
