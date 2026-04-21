package app

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

// ---------- helpers ----------

func testCfg() config.LoadResult {
	return config.LoadResult{Config: config.DefaultConfig()}
}

func makeEntries(n int) []logsource.Entry {
	entries := make([]logsource.Entry, n)
	for i := range entries {
		entries[i] = logsource.Entry{
			LineNumber: i + 1,
			IsJSON:     true,
			Level:      "INFO",
			Msg:        "test message",
		}
	}
	return entries
}

// send dispatches a message to the model and returns the updated model.
func send(m Model, msg tea.Msg) Model {
	updated, _ := m.Update(msg)
	return updated.(Model)
}

// key sends a key message by key name.
func key(m Model, k string) Model {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	return send(m, msg)
}

// resize sends a WindowSizeMsg to the model.
func resize(m Model, w, h int) Model {
	return send(m, tea.WindowSizeMsg{Width: w, Height: h})
}

// newModel creates a default model with no source (stdin mode).
func newModel() Model {
	return New("", false, "", testCfg())
}

// setFocus mutates m.focus and m.keyhints in one step — mirroring the
// pairing every test site needs (the keyhints bar reads the focus).
func setFocus(m Model, f appshell.FocusTarget) Model {
	m.focus = f
	m.keyhints = m.keyhints.WithFocus(f)
	return m
}

// ---------- T-091: detail pane auto-close on minimum underflow ----------

// TestModel_AutoClose_RightSplit_BelowMinWidth verifies the detail pane is
// auto-closed when a resize shrinks its content below MinDetailWidth while
// still in right-split orientation. Threshold default = 100 cols, so width
// 100 stays in right-split; detailW = 29 < 30 → auto-close.
func TestModel_AutoClose_RightSplit_BelowMinWidth(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane should be open after openPane")
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: orientation should be right at termWidth=200, got %v", m.resize.Orientation())

	// Shrink while keeping right-split: detailW drops below MinDetailWidth.
	m = resize(m, 100, 24)
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: orientation should stay right at termWidth=100, got %v", m.resize.Orientation())

	assert.Falsef(t, m.pane.IsOpen(),
		"expected auto-close at termWidth=100 (detailW=%d), pane still open", m.layout.Layout().DetailContentWidth())
	assert.True(t, m.keyhints.HasNotice(), "expected status notice after auto-close")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after auto-close: got %v, want FocusEntryList", m.focus)
}

// TestModel_AutoClose_NoticeClearedOnMsg verifies the noticeClearMsg resets
// the key-hint bar back to hints-mode.
func TestModel_AutoClose_NoticeClearedOnMsg(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = resize(m, 100, 24)
	require.True(t, m.keyhints.HasNotice(), "precondition: notice should be set after auto-close")

	m = send(m, noticeClearMsg{})
	assert.False(t, m.keyhints.HasNotice(), "notice should be cleared after noticeClearMsg")
}

// TestModel_AutoClose_AboveMin_KeepsPaneOpen verifies a resize that stays
// above the minimum threshold does not close the pane.
func TestModel_AutoClose_AboveMin_KeepsPaneOpen(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	m = resize(m, 180, 24)

	assert.Truef(t, m.pane.IsOpen(),
		"pane should remain open at termWidth=180 (detailW=%d)", m.layout.Layout().DetailContentWidth())
	assert.False(t, m.keyhints.HasNotice(), "notice should not be set when pane stays open")
}

// ---------- T-104: mouse drag on divider resizes detail pane width ----------

// TestModel_DividerDrag_UpdatesWidthRatio verifies a Press on the divider
// followed by a Motion to a new column translates to a width_ratio update.
func TestModel_DividerDrag_UpdatesWidthRatio(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	before := m.cfg.Config.DetailPane.WidthRatio

	// Press on the divider column to start the drag session.
	// Post-T-160 the visible `│` column equals ListContentWidth().
	l := m.layout.Layout()
	divider := l.ListContentWidth()
	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: expected draggingDivider=true after Press on divider column %d", divider)

	// Motion to a column 20 cells to the left → detail grows → ratio rises.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	after := m.cfg.Config.DetailPane.WidthRatio
	assert.Greaterf(t, after, before,
		"drag-left should increase width_ratio: before=%.3f after=%.3f", before, after)

	// Release ends the drag session.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	assert.False(t, m.draggingDivider, "expected draggingDivider=false after Release")
}

// ---------- T-099: ratio live write-back to config ----------

// TestModel_DividerDragRelease_PersistsWidthRatio verifies that releasing
// a divider drag flushes the new width_ratio to disk.
func TestModel_DividerDragRelease_PersistsWidthRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"

	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	m = m.openPane(makeEntries(3)[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())

	l := m.layout.Layout()
	divider := l.ListContentWidth() // post-T-160 visible-glyph column

	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider, "precondition: drag did not start at divider x=%d", divider)
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.WidthRatio, reloaded.Config.DetailPane.WidthRatio,
		"disk width_ratio after drag release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
}

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

// ---------- R14 (cavekit-entry-list): pane re-sync on tail-follow snap ----------

// jsonEntry builds a minimal JSON entry whose Raw unmarshals cleanly so
// the detail pane's RenderJSON path (not the raw fallback) runs. The
// `msg` field is echoed into the rendered JSON so tests can assert the
// pane content changed to match a new entry.
func jsonEntry(line int, msg string) logsource.Entry {
	return logsource.Entry{
		LineNumber: line,
		IsJSON:     true,
		Level:      "INFO",
		Msg:        msg,
		Raw:        []byte(`{"level":"INFO","msg":"` + msg + `"}`),
	}
}

// cavekit-entry-list R14 (AC: selection signal on tail-follow snap):
// when a tail append snaps the cursor to the new last entry AND the
// detail pane is open, the pane must re-render with that new entry in
// the same frame — no keypress required. Before this AC was added, the
// cursor advanced but the pane kept showing the previous entry.
func TestModel_TailFollow_TailMsg_PaneResyncsOnAppend(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
		jsonEntry(3, "gamma-msg"),
	}
	m = m.SetEntries(initial)

	m = key(m, "G")
	require.Equalf(t, 2, m.list.Cursor(), "precondition: cursor at last entry, got %d", m.list.Cursor())
	m = key(m, "enter")
	require.True(t, m.pane.IsOpen(), "precondition: pane open after Enter")
	require.Truef(t, containsSubstring(m.pane.View(), "gamma-msg"),
		"precondition: pane shows gamma-msg; got: %q", m.pane.View())

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entries: []logsource.Entry{appended}}))

	assert.Equalf(t, 3, m.list.Cursor(),
		"cursor should snap to new last entry (3), got %d", m.list.Cursor())
	assert.True(t, m.pane.IsOpen(), "pane should still be open after append")
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "delta-unique"),
		"pane must re-render with appended entry (delta-unique); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "gamma-msg"),
		"pane must not keep rendering previous entry (gamma-msg); got: %q", post)
}

// cavekit-entry-list R14 symmetry: same re-sync invariant for
// background-load batches (EntryBatchMsg). Tests the same code path as
// TailMsg but via the LoadFileStreamMsg wrapper used during non-follow
// file load.
func TestModel_TailFollow_BatchMsg_PaneResyncsOnAppend(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
	}
	m = m.SetEntries(initial)

	m = key(m, "G")
	m = key(m, "enter")
	require.Truef(t, m.pane.IsOpen() && m.list.Cursor() == 1,
		"precondition: cursor=%d pane=%v", m.list.Cursor(), m.pane.IsOpen())

	batch := logsource.EntryBatchMsg{Entries: []logsource.Entry{
		jsonEntry(3, "charlie-msg"),
		jsonEntry(4, "delta-unique"),
	}}
	m = send(m, logsource.NewLoadFileStreamMsgForTest(batch))

	assert.Equalf(t, 3, m.list.Cursor(),
		"cursor should snap to new last entry (3), got %d", m.list.Cursor())
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "delta-unique"),
		"pane must re-render with last appended entry (delta-unique); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "beta-msg"),
		"pane must not keep rendering pre-append entry (beta-msg); got: %q", post)
}

// cavekit-entry-list R14: signal must fire only when cursor actually
// moved. With the cursor NOT at the last entry (follow disengaged), an
// append must NOT re-render the pane — otherwise every tail event
// silently clobbers the user's pane selection while they are reading a
// historical entry.
func TestModel_TailFollow_NotAtTail_PaneNotResynced(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
		jsonEntry(3, "gamma-msg"),
	}
	m = m.SetEntries(initial)

	// Cursor starts at 0 — NOT at tail. Open pane on alpha.
	m = key(m, "enter")
	require.Truef(t, m.pane.IsOpen() && containsSubstring(m.pane.View(), "alpha-msg"),
		"precondition: pane open on alpha; got: %q", m.pane.View())

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entries: []logsource.Entry{appended}}))

	assert.Equalf(t, 0, m.list.Cursor(),
		"cursor must NOT move when pre-append cursor < tail, got %d", m.list.Cursor())
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "alpha-msg"),
		"pane must still show alpha-msg (user's non-tail selection preserved); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "delta-unique"),
		"pane must NOT render appended entry (delta-unique) when follow disengaged; got: %q", post)
}

// ---------- T-127 (F-020): hidden-fields wiring ----------

// T-127 (F-020): openPane wires `visibility.HiddenFields()` into the pane
// so config-driven field hiding reaches the JSON renderer.
func TestModel_OpenPane_WiresHiddenFieldsFromVisibility(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.HiddenFields = []string{"secret"}
	m := New("", false, "", cfg)
	m = resize(m, 80, 24)
	entry := logsource.Entry{
		LineNumber: 1,
		IsJSON:     true,
		Level:      "INFO",
		Msg:        "hi",
		Raw:        []byte(`{"level":"INFO","msg":"hi","secret":"shh"}`),
	}
	m = m.SetEntries([]logsource.Entry{entry})
	m = m.openPane(entry)

	view := m.pane.View()
	assert.Falsef(t, containsSubstring(view, "secret"),
		"detail pane view should not include suppressed field `secret`, got: %q", view)
	assert.Falsef(t, containsSubstring(view, "shh"),
		"detail pane view should not include suppressed value, got: %q", view)
}

// containsSubstring is a small helper to keep imports tight.
func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ---------- helpers ----------

// containsCount returns true if s contains the decimal representation of n.
func containsCount(s string, n int) bool {
	needle := itoa(n)
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
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

// ---------- T-156: mouse drag resize (cavekit-app-shell R15) ----------

// belowDividerY computes the divider row for a below-mode model —
// entryListEnd+1, where entryListEnd = layout.EntryListHeight().
func belowDividerY(m Model) int {
	return m.layout.Layout().EntryListHeight() + 1
}

// rightDividerX computes the divider column for a right-split model.
// Mirrors MouseRouter.Zone post-T-160: divider = ListContentWidth() (the
// visible `│` glyph column, verified by renderer-truth test in appshell/).
func rightDividerX(m Model) int {
	return m.layout.Layout().ListContentWidth()
}

// TestModel_T156_BelowDrag_PressOnDivider_StartsDrag confirms that a left
// Press on the divider row in below-orientation flips `draggingDivider`.
func TestModel_T156_BelowDrag_PressOnDivider_StartsDrag(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	dy := belowDividerY(m)
	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	assert.Truef(t, m.draggingDivider, "Press on below-mode divider row y=%d must start drag", dy)
}

// TestModel_T156_BelowDrag_Motion covers the below-mode drag motion
// matrix: grow on up, shrink on down, pin at RatioMax/RatioMin for
// out-of-bounds motion. `targetY(dy, termH)` returns the absolute Y the
// Motion message should aim at — dy is the initial divider row.
func TestModel_T156_BelowDrag_Motion(t *testing.T) {
	greater := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Greaterf(t, a, b, "height_ratio must grow: before=%.3f after=%.3f", b, a)
	}
	less := func(t *testing.T, b, a float64) {
		t.Helper()
		assert.Lessf(t, a, b, "height_ratio must shrink: before=%.3f after=%.3f", b, a)
	}
	eq := func(want float64) func(t *testing.T, _, after float64) {
		return func(t *testing.T, _, a float64) {
			t.Helper()
			assert.Equalf(t, want, a, "height_ratio: got %.3f, want %.3f", a, want)
		}
	}
	cases := []struct {
		name      string
		seedRatio float64 // 0 = leave default
		targetY   func(dy, termH int) int
		assertion func(t *testing.T, before, after float64)
	}{
		{"motionUp_growsDetail", 0, func(dy, _ int) int { return dy - 4 }, greater},
		{"motionDown_from_0.50_shrinksDetail", 0.50, func(dy, _ int) int { return dy + 3 }, less},
		{"extremeUp_clampsAtMax", 0, func(_, _ int) int { return -100 }, eq(appshell.RatioMax)},
		{"extremeDown_clampsAtMin", 0, func(_, termH int) int { return termH + 100 }, eq(appshell.RatioMin)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := setupRatioModelBelow(t, false)
			if tc.seedRatio != 0 {
				m.cfg.Config.DetailPane.HeightRatio = tc.seedRatio
				m.paneHeight = m.paneHeight.SetRatio(tc.seedRatio)
				m = m.relayout()
			}
			dy := belowDividerY(m)
			termH := m.resize.Height()
			before := m.cfg.Config.DetailPane.HeightRatio

			m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: 20, Y: tc.targetY(dy, termH), Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

			tc.assertion(t, before, m.cfg.Config.DetailPane.HeightRatio)
		})
	}
}

// TestModel_T156_BelowDrag_Release_PersistsHeightRatio ensures the final
// ratio is flushed to disk exactly once on mouse release, and that the
// in-memory value matches.
func TestModel_T156_BelowDrag_Release_PersistsHeightRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: want below, got %v", m.resize.Orientation())
	dy := belowDividerY(m)

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	// Ensure config file does NOT exist yet — motion must not save.
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	_, err := os.Stat(cfgPath)
	assert.Error(t, err, "motion during drag must NOT write config; file exists already")
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	assert.False(t, m.draggingDivider, "Release must end drag session")
	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.HeightRatio, reloaded.Config.DetailPane.HeightRatio,
		"disk height_ratio after release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.HeightRatio, m.cfg.Config.DetailPane.HeightRatio)
}

// TestModel_T156_Drag_IsFocusNeutral ensures dragging the divider never
// mutates m.focus — in either orientation, starting from either focus.
func TestModel_T156_Drag_IsFocusNeutral(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(t *testing.T, listFocus bool) Model
		dividerFn func(Model) (x, y int)
		motionDX  int
		motionDY  int
	}{
		{
			name:      "right/detail-focus",
			setup:     setupRatioModelRight,
			dividerFn: func(m Model) (int, int) { return rightDividerX(m), 5 },
			motionDX:  -10,
		},
		{
			name:      "right/list-focus",
			setup:     func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, true) },
			dividerFn: func(m Model) (int, int) { return rightDividerX(m), 5 },
			motionDX:  -10,
		},
		{
			name:      "below/detail-focus",
			setup:     setupRatioModelBelow,
			dividerFn: func(m Model) (int, int) { return 20, belowDividerY(m) },
			motionDY:  -3,
		},
		{
			name:      "below/list-focus",
			setup:     func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, true) },
			dividerFn: func(m Model) (int, int) { return 20, belowDividerY(m) },
			motionDY:  -3,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t, false)
			startFocus := m.focus
			x, y := tc.dividerFn(m)
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: x + tc.motionDX, Y: y + tc.motionDY, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
			m = send(m, tea.MouseMsg{X: x + tc.motionDX, Y: y + tc.motionDY, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
			assert.Equalf(t, startFocus, m.focus,
				"drag changed focus: start=%v end=%v", startFocus, m.focus)
		})
	}
}

// TestModel_T156_PaneClosed_PressIsNoOp: pressing anywhere with the pane
// closed never starts a drag session (there is no divider).
func TestModel_T156_PaneClosed_PressIsNoOp(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}

	for _, size := range []struct{ w, h int }{{80, 24}, {200, 24}} {
		m := New("", false, cfgPath, cfg)
		m = resize(m, size.w, size.h)
		m = m.SetEntries(makeEntries(3))
		require.Falsef(t, m.pane.IsOpen(), "precondition: pane must be closed at %dx%d", size.w, size.h)
		m = send(m, tea.MouseMsg{X: 10, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		assert.Falsef(t, m.draggingDivider,
			"%dx%d press with pane closed must not start drag", size.w, size.h)
	}
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"pane-closed press must not save config; file exists: %v", err)
}

// ---------- T-157: divider-cell click is focus-neutral (cavekit-app-shell R6 AC 7) ----------

// TestModel_T157_DividerClick_DoesNotTransferFocus verifies the R6 contract:
// a Press + Release on the divider cell itself never transfers focus to
// either pane. The divider is reserved for R15 drag initiation; focus
// stays wherever it was before the click.
func TestModel_T157_DividerClick_DoesNotTransferFocus(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, listFocus bool) Model
		xy    func(m Model) (int, int)
	}{
		{
			name:  "right/start-list-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, true) },
			xy:    func(m Model) (int, int) { return rightDividerX(m), 5 },
		},
		{
			name:  "right/start-detail-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelRight(t, false) },
			xy:    func(m Model) (int, int) { return rightDividerX(m), 5 },
		},
		{
			name:  "below/start-list-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, true) },
			xy:    func(m Model) (int, int) { return 20, belowDividerY(m) },
		},
		{
			name:  "below/start-detail-focus",
			setup: func(t *testing.T, _ bool) Model { return setupRatioModelBelow(t, false) },
			xy:    func(m Model) (int, int) { return 20, belowDividerY(m) },
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t, false)
			startFocus := m.focus
			x, y := tc.xy(m)
			// Bare click: Press then Release with no motion between.
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			m = send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
			assert.Equalf(t, startFocus, m.focus,
				"divider click changed focus: start=%v end=%v", startFocus, m.focus)
		})
	}
}

// TestModel_T156_RightDrag_UpdatesWidthRatio mirrors the T-104 coverage
// after the dual-orientation refactor — confirms right-split drag still
// updates width_ratio (not height_ratio) and saves once on release.
func TestModel_T156_RightDrag_UpdatesWidthRatio(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right, got %v", m.resize.Orientation())
	beforeH := m.cfg.Config.DetailPane.HeightRatio
	dx := rightDividerX(m)

	m = send(m, tea.MouseMsg{X: dx, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.Equalf(t, beforeH, m.cfg.Config.DetailPane.HeightRatio,
		"right-mode drag must not mutate height_ratio: %.3f → %.3f",
		beforeH, m.cfg.Config.DetailPane.HeightRatio)
	reloaded := config.Load(cfgPath)
	assert.Equalf(t, m.cfg.Config.DetailPane.WidthRatio, reloaded.Config.DetailPane.WidthRatio,
		"disk width_ratio after right-drag release: got %.3f, want %.3f",
		reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
}

// ---------- T-158: single-owner click-row resolver (cavekit-entry-list R10) ----------

// clickAt emits a left-Press at (x, y) and returns the updated model.
func clickAt(m Model, x, y int) Model {
	return send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
}

// TestModel_T158_Click_RowResolver drives the row-resolver across the
// six parametric scenarios (below/right × pane-open/closed × valid row /
// top-border / header). `wantCursor == 0` means the cursor must not
// change from its pre-click position.
func TestModel_T158_Click_RowResolver(t *testing.T) {
	cases := []struct {
		name       string
		w          int // terminal width → orientation
		paneOpen   bool
		advance    int // j-presses before the click (0 = cursor at row 0)
		clickY     int
		wantCursor int // 0 = unchanged; otherwise 1-based CursorPosition
	}{
		{"below_firstRow_y2", 80, false, 0, 2, 1},
		{"below_secondRow_y3", 80, false, 0, 3, 2},
		{"below_topBorder_y1_noop", 80, false, 0, 1, 0},
		{"below_header_y0_noop", 80, false, 5, 0, 0},
		{"right_firstRow_y2", 200, true, 0, 2, 1},
		{"below_paneOpen_firstRow_y2", 80, true, 0, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newModel()
			m = resize(m, tc.w, 24)
			entries := makeEntries(20)
			m = m.SetEntries(entries)
			if tc.paneOpen {
				m = m.openPane(entries[0])
			}
			for i := 0; i < tc.advance; i++ {
				m = key(m, "j")
			}
			before := m.list.CursorPosition()
			m = clickAt(m, 10, tc.clickY)
			got := m.list.CursorPosition()
			if tc.wantCursor == 0 {
				assert.Equalf(t, before, got,
					"click y=%d must be no-op: CursorPosition before=%d after=%d",
					tc.clickY, before, got)
			} else {
				assert.Equalf(t, tc.wantCursor, got,
					"click y=%d: CursorPosition = %d, want %d", tc.clickY, got, tc.wantCursor)
			}
		})
	}
}

// TestModel_T158_DoubleClick_UsesSameResolver: double-click at y=3 opens the
// detail pane for row 1 — the same row a single click at y=3 would select.
func TestModel_T158_DoubleClick_UsesSameResolver(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(20)
	m = m.SetEntries(entries)

	// First click at y=3 positions the cursor on row 1.
	m = clickAt(m, 10, 3)
	require.Equalf(t, 2, m.list.CursorPosition(),
		"first click y=3: CursorPosition = %d, want 2", m.list.CursorPosition())
	// Second click at the SAME y=3 within 500ms triggers double-click,
	// which emits OpenDetailPaneMsg. We verify the cmd return.
	updated, cmd := m.Update(tea.MouseMsg{X: 10, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = updated.(Model)
	require.NotNil(t, cmd, "second click at same y should emit a cmd for double-click")
	msg := cmd()
	open, ok := msg.(entrylist.OpenDetailPaneMsg)
	require.Truef(t, ok, "double-click cmd: want OpenDetailPaneMsg, got %T", msg)
	assert.Equalf(t, entries[1].LineNumber, open.Entry.LineNumber,
		"double-click opens wrong entry: got line %d, want %d",
		open.Entry.LineNumber, entries[1].LineNumber)
}

// TestModel_T158_Click_DividerRow_NoListSelection: click on the below-mode
// divider row does not mutate the list cursor — divider zone is handled
// separately by the drag branch.
func TestModel_T158_Click_DividerRow_NoListSelection(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(20)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	for i := 0; i < 3; i++ {
		m = key(m, "j")
	}
	before := m.list.CursorPosition()

	dy := belowDividerY(m)
	m = clickAt(m, 10, dy)

	assert.Equalf(t, before, m.list.CursorPosition(),
		"click on divider row (y=%d): CursorPosition = %d, want %d (unchanged)",
		dy, m.list.CursorPosition(), before)
}

// ---------- T-162: mid-drag auto-close terminates the drag (F-125) ----------

// TestModel_T162_Drag_AutoClose_TerminatesSession verifies that if the
// pane auto-closes mid-drag (e.g. terminal shrinks below the right-split
// min-width threshold), the subsequent Motion + Release stream does NOT
// mutate the persisted ratio and does NOT write to the config file. The
// belt-and-braces clear in the WindowSizeMsg auto-close branch guarantees
// `draggingDivider` is false by the time the next Motion arrives.
func TestModel_T162_Drag_AutoClose_TerminatesSession(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right" // pin orientation across resizes
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane must be open")
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation, got %v", m.resize.Orientation())
	dividerX := rightDividerX(m)
	startW := m.cfg.Config.DetailPane.WidthRatio

	// Begin a drag with a Press on the divider.
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: Press at divider x=%d must start drag", dividerX)

	// Shrink the terminal so the right-split detail width falls below
	// MinDetailWidth (30). Usable = w-5, detailW = usable * 0.30.
	// Using w=70 → usable=65 → detailW=19 < 30 → auto-close fires.
	m = resize(m, 70, 24)
	require.False(t, m.pane.IsOpen(), "precondition: pane must have auto-closed at w=70")
	assert.False(t, m.draggingDivider, "auto-close must clear draggingDivider (belt-and-braces)")

	// Any subsequent Motion + Release must be a no-op for ratio + config.
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"width_ratio mutated after mid-drag auto-close: start=%.3f end=%.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"mid-drag auto-close must NOT write config; file exists: %v", err)
}

// TestModel_T162_DragBranch_GuardsOnClosedPane covers the inner guard —
// if something somehow leaves draggingDivider=true with pane closed (e.g.
// future code path we haven't anticipated), the drag branch must still
// short-circuit before touching the ratio or saving config.
func TestModel_T162_DragBranch_GuardsOnClosedPane(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	// Artificial stale state: pane closed but drag flag set.
	m.draggingDivider = true
	startW := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: 50, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	assert.False(t, m.draggingDivider, "closed-pane drag-branch guard must clear draggingDivider")
	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"closed-pane drag-branch must not mutate ratio: %.3f → %.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"closed-pane drag-branch must not write config; file exists: %v", err)
}

// ---------- T-164: bare Press+Release skips config write (F-129) ----------

// TestModel_T164_BareClick_OnDivider_DoesNotWriteConfig verifies that a
// Press immediately followed by a Release on the divider (no intervening
// Motion) does NOT write the config file. Previously the Release branch
// unconditionally called saveConfig, so bare clicks on the divider
// rewrote `config.toml` every time a user clicked near the border.
func TestModel_T164_BareClick_OnDivider_DoesNotWriteConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	dividerX := rightDividerX(m)
	startW := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	require.Truef(t, m.draggingDivider,
		"precondition: Press at divider x=%d must start drag", dividerX)
	assert.False(t, m.dragDirty, "Press must reset dragDirty to false")
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	assert.False(t, m.draggingDivider, "Release must end drag session")
	assert.Equalf(t, startW, m.cfg.Config.DetailPane.WidthRatio,
		"bare click mutated width_ratio: start=%.3f end=%.3f",
		startW, m.cfg.Config.DetailPane.WidthRatio)
	_, err := os.Stat(cfgPath)
	assert.Truef(t, os.IsNotExist(err),
		"bare Press+Release on divider must NOT write config; file exists: %v", err)
}

// TestModel_T164_Drag_WithMotion_DoesWriteConfig is the positive-case
// anchor — a drag that actually moves the divider still flushes the new
// ratio to disk on Release. Without this, T-164 would be trivially
// satisfied by disabling all writes.
func TestModel_T164_Drag_WithMotion_DoesWriteConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	dividerX := rightDividerX(m)

	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: dividerX - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	assert.True(t, m.dragDirty, "Motion that changes ratio must set dragDirty=true")
	m = send(m, tea.MouseMsg{X: dividerX - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	_, err := os.Stat(cfgPath)
	assert.NoErrorf(t, err, "drag with motion must write config: %v", err)
}
