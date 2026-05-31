package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// V42 (B29): View() is pure — the per-pane Focused/Alone flags are written in
// Update (syncPaneState), never at compose time. Two consecutive View() calls
// MUST be byte-equal, and the flags must already reflect the model state after
// Update (not be recomputed inside View).
func TestModel_V42_ViewPure_FlagsSetInUpdate_B29(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)
	m = m.SetEntries(makeEntries(5))

	// Pane open, focus stays on list (V12 — open does not transfer focus).
	m = key(m, "enter")
	require.True(t, m.pane.IsOpen(), "precondition: pane open after Enter")
	assert.True(t, m.list.Focused, "list focused flag set in Update")
	assert.False(t, m.list.Alone, "list not Alone while pane open")
	assert.False(t, m.pane.Focused, "pane unfocused while focus on list")

	// Determinism: View() must not mutate state ⇒ identical bytes across calls.
	v1 := m.View()
	v2 := m.View()
	assert.Equal(t, v1, v2, "View() must be deterministic (no compose-time writes)")

	// Tab transfers focus to the pane; Update re-syncs the flags.
	m = key(m, "tab")
	require.Equal(t, appshell.FocusDetailPane, m.focus, "Tab focuses the pane")
	assert.False(t, m.list.Focused, "list unfocused after Tab to pane")
	assert.True(t, m.pane.Focused, "pane focused flag set in Update after Tab")

	v3 := m.View()
	v4 := m.View()
	assert.Equal(t, v3, v4, "View() still deterministic with pane focused")
}

// V42 (B30): m.focus must always reference a VISIBLE pane. If the detail pane
// closes while focus is on it (auto-close, drag-collapse, ratio-min, explicit),
// the next Update re-homes focus to a visible peer in the same step — so the
// following key routes to the list, not into the void.
func TestModel_V42_FocusReHomesOnPaneClose_B30(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "enter")
	m = key(m, "tab") // focus the pane
	require.Equal(t, appshell.FocusDetailPane, m.focus, "precondition: focus on pane")

	// Simulate a close path that forgets to re-sync focus (the B30 failure
	// mode: e.g. a future drag-collapse). The global guard must fix it.
	m.pane = m.pane.Close()
	require.False(t, m.pane.IsOpen(), "precondition: pane closed under focus")

	// Any subsequent message runs syncPaneState.
	m = send(m, noticeClearMsg{})
	assert.Equal(t, appshell.FocusEntryList, m.focus,
		"focus must re-home to the visible list when the pane closed under it")
	assert.True(t, m.list.Focused, "list focused flag tracks the re-homed focus")

	// And a key now routes to the list (cursor moves), not to a hidden pane.
	before := m.list.Cursor()
	m = key(m, "j")
	assert.Equalf(t, before+1, m.list.Cursor(),
		"j must route to the list after focus re-home (was %d)", before)
}

// V42 (B31): an active in-pane search owns the pane's scroll + query. A
// tail-follow append that snaps the list cursor must NOT clobber that state by
// re-opening the pane on the new entry — the query survives and the pane keeps
// rendering the searched entry (scroll preserved, since Open is what resets it).
func TestModel_V42_PaneSearchSurvivesAppend_B31(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)
	m = m.SetEntries([]logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
		jsonEntry(3, "gamma-original"),
	})
	m = key(m, "G") // cursor to tail
	m = key(m, "enter")
	require.True(t, m.pane.IsOpen(), "precondition: pane open on tail entry")
	require.Truef(t, containsSubstring(m.pane.View(), "gamma-original"),
		"precondition: pane shows gamma-original; got %q", m.pane.View())

	// Focus the pane and open a search with a non-empty query.
	m = setFocus(m, appshell.FocusDetailPane)
	m.paneSearch = m.paneSearch.Activate().SetQuery("gamma", m.pane.ContentLines())
	require.True(t, m.paneSearch.IsActive(), "precondition: pane search active")
	require.Equal(t, "gamma", m.paneSearch.Query(), "precondition: query set")

	// Tail append snaps the list cursor to the new last entry.
	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entries: []logsource.Entry{appended}}))

	assert.Equal(t, 3, m.list.Cursor(), "list cursor still snaps to tail")
	assert.True(t, m.paneSearch.IsActive(), "pane search must survive the append")
	assert.Equal(t, "gamma", m.paneSearch.Query(), "query must survive the append")

	post := m.pane.View()
	// "gamma" is wrapped in search-highlight SGR codes, so assert on the
	// un-highlighted tail of the searched entry's msg.
	assert.Truef(t, containsSubstring(post, "original"),
		"pane must keep rendering the searched entry; got %q", post)
	assert.Falsef(t, containsSubstring(post, "delta-unique"),
		"pane must NOT re-open on the appended entry while search active; got %q", post)
}
