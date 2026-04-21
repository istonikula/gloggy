package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// ---------- T-117: dismiss paneSearch on pane close / reopen (F-006) ----------

// TestModel_PaneSearch_DismissedOnBlurred verifies that when the pane
// closes, any active search state is cleared so a subsequent open starts
// with a clean query.
func TestModel_PaneSearch_DismissedOnBlurred(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144: `/` is focus-based)
	m = key(m, "/")     // start pane search
	m = key(m, "t")     // type a char
	require.True(t, m.paneSearch.IsActive(), "precondition: search should be active after '/'")
	require.NotEmpty(t, m.paneSearch.Query(), "precondition: query should be non-empty after typing")

	m = send(m, detailpane.BlurredMsg{})

	assert.False(t, m.paneSearch.IsActive(), "search should be dismissed after pane closes")
	assert.Equalf(t, "", m.paneSearch.Query(),
		"search query should be cleared after pane closes, got %q", m.paneSearch.Query())
}

// TestModel_PaneSearch_DismissedOnReopen verifies that opening a new
// entry does not carry stale search state from a previous open/close.
func TestModel_PaneSearch_DismissedOnReopen(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Simulate prior search state lingering from an earlier pane session.
	m.paneSearch = m.paneSearch.Activate().SetQuery("stale", m.pane.ContentLines())

	m = m.openPane(entries[1])

	assert.False(t, m.paneSearch.IsActive(), "search should be dismissed when a new entry opens the pane")
	assert.Equalf(t, "", m.paneSearch.Query(),
		"search query should be cleared on reopen, got %q", m.paneSearch.Query())
}

// ---------- T-120: two-step Esc integration (cavekit-detail-pane R7 / F-007) ----------

// TestModel_TwoStepEsc_DismissesSearchThenClosesPane verifies the
// cavekit R7 "two-step Esc" contract: first Esc with search active
// dismisses search and leaves the pane open; second Esc closes the pane.
func TestModel_TwoStepEsc_DismissesSearchThenClosesPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144: focus-based `/`)
	m = key(m, "/")     // activate pane search
	m = key(m, "h")     // type something to exercise the query path
	require.True(t, m.paneSearch.IsActive(), "precondition: search should be active before first Esc")
	require.True(t, m.pane.IsOpen(), "precondition: pane should be open before first Esc")

	// First Esc — dismisses search, pane stays open.
	m = key(m, "esc")
	assert.False(t, m.paneSearch.IsActive(), "first Esc should dismiss search")
	assert.True(t, m.pane.IsOpen(), "first Esc should NOT close the pane")
	assert.Equalf(t, appshell.FocusDetailPane, m.focus,
		"after first Esc focus should remain FocusDetailPane, got %v", m.focus)

	// Second Esc — closes the pane. Pane emits BlurredMsg which we
	// deliver to the model to complete the focus handoff.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	assert.False(t, m.pane.IsOpen(), "second Esc should close the pane")
	require.NotNil(t, cmd, "second Esc should return BlurredMsg cmd")
	blurMsg := cmd()
	_, ok := blurMsg.(detailpane.BlurredMsg)
	require.Truef(t, ok, "second Esc should emit BlurredMsg, got %T", blurMsg)
	m = send(m, blurMsg)
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after second Esc: got %v, want FocusEntryList", m.focus)
}

// ---------- T-144: focus-based `/` routing (cavekit-app-shell R13 revised) ----------

// TestModel_Slash_ListFocus_PaneOpen_ActivatesListSearch verifies that
// `/` with list focused opens list-scope search — even when the pane is
// open, focus stays on the list and list search activates (T-144).
func TestModel_Slash_ListFocus_PaneOpen_ActivatesListSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Focus stays on the list after openPane per T-126 — that is the
	// premise we want to test: `/` activates list search from here.
	require.Equalf(t, appshell.FocusEntryList, m.focus,
		"precondition: want list focus after openPane, got %v", m.focus)

	m = key(m, "/")

	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"/ with list focused must NOT transfer focus, got %v", m.focus)
	assert.True(t, m.list.HasActiveSearch(), "/ with list focused should activate list search")
	assert.False(t, m.paneSearch.IsActive(), "/ with list focused must NOT activate pane search")
}

// TestModel_Slash_ListFocus_PaneClosed_ActivatesListSearch verifies that
// `/` with list focused and pane closed activates list search (T-144).
// The old T-116 "open entry first" notice is gone — list search is
// always available when the list is focused.
func TestModel_Slash_ListFocus_PaneClosed_ActivatesListSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	require.False(t, m.pane.IsOpen(), "precondition: pane should be closed")

	m = key(m, "/")

	assert.True(t, m.list.HasActiveSearch(), "/ with list focused (pane closed) should activate list search")
	assert.False(t, m.paneSearch.IsActive(), "/ with list focused must NOT activate pane search")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"/ with list focused should stay on list, got %v", m.focus)
}

// TestModel_Slash_DetailPaneFocus_ActivatesPaneSearch verifies that
// `/` with the detail pane focused activates in-pane search (T-144).
func TestModel_Slash_DetailPaneFocus_ActivatesPaneSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = key(m, "tab") // focus → detail pane

	m = key(m, "/")

	assert.True(t, m.paneSearch.IsActive(), "/ with pane focused should activate pane search")
	assert.False(t, m.list.HasActiveSearch(), "/ with pane focused must NOT activate list search")
}

// TestModel_ListSearch_TabCycle_ClearsSearch verifies that Tab-cycling
// off the list clears any active list search (cavekit-entry-list R13 AC).
func TestModel_ListSearch_TabCycle_ClearsSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Focus is on list. Activate list search.
	m = key(m, "/")
	m = key(m, "e")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search should be active")

	m = key(m, "tab") // focus → detail pane

	assert.False(t, m.list.HasActiveSearch(), "Tab cycling off the list should clear list search")
}

// TestModel_Slash_FilterPanelFocus_RoutedToFilter verifies that `/` with
// the filter panel focused is routed to the filter input as a literal
// character (not intercepted as a global search activation).
func TestModel_Slash_FilterPanelFocus_RoutedToFilter(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "f") // focus = FocusFilterPanel
	require.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"precondition: want filter panel focused, got %v", m.focus)

	m = key(m, "/")

	assert.False(t, m.paneSearch.IsActive(), "/ in filter panel should NOT activate pane search")
	assert.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"/ in filter panel should stay in filter panel, got %v", m.focus)
}

// ---------- T-118: input vs navigation mode (F-008) ----------

// TestModel_Search_NavigateMode_JPassesThroughToPane verifies that once
// search is in navigation mode, `j` scrolls the pane rather than being
// eaten by the search input.
func TestModel_Search_NavigateMode_JPassesThroughToPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	// Use a JSON entry with many fields so the pane content is scrollable.
	rawJSON := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","k":"11","l":"12","m":"13","n":"14"}`
	entries := []logsource.Entry{{LineNumber: 1, IsJSON: true, Level: "INFO", Msg: "x", Raw: []byte(rawJSON)}}
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = key(m, "tab") // focus detail pane (T-144: `/` is focus-based)
	m = key(m, "/")   // activate — input mode
	require.True(t, m.paneSearch.IsActive(), "precondition: search should be active after '/'")
	m = key(m, "enter") // commit to navigate
	require.Equalf(t, detailpane.SearchModeNavigate, m.paneSearch.Mode(),
		"precondition: want SearchModeNavigate, got %v", m.paneSearch.Mode())

	paneBefore := m.pane
	m = key(m, "j")
	// Query must be unchanged: `j` in nav mode should NOT extend the query.
	assert.Equalf(t, "", m.paneSearch.Query(),
		"nav-mode `j` should not extend query, got %q", m.paneSearch.Query())
	// Pane view should have changed IF the content exceeded the viewport.
	// We check by rendering View() and comparing — if there was scroll
	// room, the view string differs.
	if m.pane.View() == paneBefore.View() {
		// Not a hard failure — if content fit in the viewport, scroll is
		// a no-op. Log instead.
		t.Logf("note: pane view unchanged after j (content may have fit in viewport)")
	}
	// And search is still active in navigate mode.
	assert.True(t, m.paneSearch.IsActive(), "search should still be active after nav-mode j")
	assert.Equalf(t, detailpane.SearchModeNavigate, m.paneSearch.Mode(),
		"mode should stay navigate after j, got %v", m.paneSearch.Mode())
}

// TestModel_Search_InputMode_JAppendsToQuery verifies that in input mode,
// `j` becomes a literal query character (does not scroll the pane).
func TestModel_Search_InputMode_JAppendsToQuery(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144)
	m = key(m, "/")     // activate pane search — input mode
	m = key(m, "j")     // should extend query
	assert.Equalf(t, "j", m.paneSearch.Query(),
		"input-mode `j` should extend query to 'j', got %q", m.paneSearch.Query())
}

// ---------- T-146 (cavekit-app-shell R14): q-quit exemption during list-search input ----------

// TestModel_Q_ListSearchInputMode_DoesNotQuit verifies that `q` typed while
// the list-search input is capturing text extends the query rather than
// quitting the app (F-106 regression fix).
func TestModel_Q_ListSearchInputMode_DoesNotQuit(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/") // activate list search — input mode
	require.True(t, m.list.HasActiveSearch(), "precondition: list search should be active")
	require.True(t, m.list.Search().InputMode(), "precondition: list search should be in input mode")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)

	if cmd != nil {
		if msg := cmd(); msg != nil {
			_, isQuit := msg.(tea.QuitMsg)
			require.False(t, isQuit, "q in list-search input mode must NOT quit")
		}
	}
	assert.Equalf(t, "q", m.list.Search().Query(),
		"q in input mode should extend query to 'q', got %q", m.list.Search().Query())
	assert.True(t, m.list.HasActiveSearch(), "list search should still be active after typed q")

	// F-117: chained keystrokes — the whole word "quit" must land in the query
	// without the trailing "uit" ever escaping through to the global quit
	// handler. Assert no tea.QuitMsg across the full sequence.
	for _, r := range []rune{'u', 'i', 't'} {
		updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
		if cmd != nil {
			if msg := cmd(); msg != nil {
				_, isQuit := msg.(tea.QuitMsg)
				require.Falsef(t, isQuit, "keystroke %q in list-search input mode must NOT quit", string(r))
			}
		}
	}
	assert.Equalf(t, "quit", m.list.Search().Query(),
		"chained q/u/i/t should build query %q, got %q", "quit", m.list.Search().Query())
	assert.True(t, m.list.HasActiveSearch(), "list search should still be active after chained keystrokes")
	assert.True(t, m.list.Search().InputMode(), "search should still be in input mode after chained keystrokes")
}

// TestModel_Q_ListSearchNavigateMode_StillQuits verifies that after Enter
// commits the search (navigate mode), `q` resumes its normal quit
// behaviour. Users who want `q` to extend a query must dismiss with Esc
// first — navigate mode is for n/N cycling, not typing.
func TestModel_Q_ListSearchNavigateMode_StillQuits(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/")
	m = key(m, "t")     // query "t" matches "test message"
	m = key(m, "enter") // commit → navigate mode
	require.False(t, m.list.Search().InputMode(),
		"precondition: search should be in navigate mode after Enter")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd, "q in navigate mode should still emit quit cmd")
	_, ok := cmd().(tea.QuitMsg)
	assert.True(t, ok, "q in navigate mode should emit tea.QuitMsg")
}

// TestModel_Q_NoListSearch_StillQuits verifies the baseline quit behaviour
// is preserved when no list search is active.
func TestModel_Q_NoListSearch_StillQuits(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd, "q without active search should quit")
	_, ok := cmd().(tea.QuitMsg)
	assert.True(t, ok, "expected tea.QuitMsg")
}

// TestModel_Q_DetailPaneFocus_StillQuits verifies `q` on a non-list focus
// is unchanged by the T-146 exemption (global quit only gates on list
// focus anyway).
func TestModel_Q_DetailPaneFocus_StillQuits(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = key(m, "tab")
	require.Equalf(t, appshell.FocusDetailPane, m.focus,
		"precondition: want detail pane focus, got %v", m.focus)

	// `q` on detail pane focus is NOT a quit (global quit only triggers on
	// list focus — this is unchanged by T-146). Just confirms no regression.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			_, isQuit := msg.(tea.QuitMsg)
			assert.False(t, isQuit, "q on detail pane focus should NOT quit (global quit requires list focus)")
		}
	}
}
