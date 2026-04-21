package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

// ---------- focus transitions ----------

// TestModel_F_OpensFocusOnFilterPanel verifies 'f' from the entry list moves
// focus to the filter panel.
func TestModel_F_OpensFocusOnFilterPanel(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = key(m, "f")

	assert.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"focus after 'f': got %v, want FocusFilterPanel", m.focus)
}

// TestModel_Enter_OpensDetailPane verifies pressing Enter when entries exist
// opens the detail pane. T-126 (F-017): focus stays on the entry list so
// `j`/`k` keep navigating the list with the pane as a live preview.
func TestModel_Enter_OpensDetailPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "enter")

	assert.True(t, m.pane.IsOpen(), "pane should be open after Enter")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after Enter: got %v, want FocusEntryList (pane open does NOT transfer focus)", m.focus)
}

// TestModel_Enter_NoEntries_DoesNotOpenPane verifies Enter with no entries
// does not open the pane.
func TestModel_Enter_NoEntries_DoesNotOpenPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	// No entries loaded.

	m = key(m, "enter")

	assert.False(t, m.pane.IsOpen(), "pane should NOT open when there are no entries")
}

// TestModel_BlurredMsg_ReturnsFocusToList verifies that detailpane.BlurredMsg
// returns focus to the entry list.
//
// In normal usage, BlurredMsg is emitted by the pane after it closes itself
// (on Esc/Enter). By the time the parent receives it, m.pane is already closed.
// This test checks only the focus transition.
func TestModel_BlurredMsg_ReturnsFocusToList(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	// Tab to the pane to mirror the path where BlurredMsg originates.
	m = key(m, "tab")

	require.Equal(t, appshell.FocusDetailPane, m.focus, "expected FocusDetailPane before sending BlurredMsg")

	m = send(m, detailpane.BlurredMsg{})

	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after BlurredMsg: got %v, want FocusEntryList", m.focus)
}

// TestModel_EscInPane_EmitsBlurredMsg verifies Esc from within the detail pane
// emits a detailpane.BlurredMsg command for the parent to process.
func TestModel_EscInPane_EmitsBlurredMsg(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // move focus to pane before testing pane-local Esc

	// Esc in the detail pane emits BlurredMsg asynchronously via a returned cmd.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	require.NotNil(t, cmd, "Esc in detail pane should return a non-nil cmd")
	blurMsg := cmd()
	_, ok := blurMsg.(detailpane.BlurredMsg)
	require.Truef(t, ok, "expected detailpane.BlurredMsg from Esc, got %T", blurMsg)

	// Deliver the BlurredMsg to the parent model to complete the focus handoff.
	m = send(m, blurMsg)
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after BlurredMsg: got %v, want FocusEntryList", m.focus)
}

// ---------- help overlay ----------

// TestModel_HelpOverlay_InterceptsKeys verifies that while the help overlay is open,
// key events do not propagate to the entry list (cursor should not move).
func TestModel_HelpOverlay_InterceptsKeys(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))

	cursorBefore := m.list.Cursor()

	// Open help overlay.
	m = key(m, "?")
	require.True(t, m.help.IsOpen(), "help overlay should be open after '?'")

	// Press 'j' — should be intercepted by the overlay and NOT move the cursor.
	m = key(m, "j")
	assert.Equalf(t, cursorBefore, m.list.Cursor(),
		"cursor moved while help overlay was open: before=%d after=%d",
		cursorBefore, m.list.Cursor())
}

// ---------- T-096: Tab focus cycle ----------

// TestModel_Tab_CyclesListToDetails verifies Tab flips list → details when
// both panes are visible. T-126: after Enter the focus stays on the list,
// so no manual reset is needed before Tab.
func TestModel_Tab_CyclesListToDetails(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)

	m = key(m, "tab")
	assert.Equalf(t, appshell.FocusDetailPane, m.focus,
		"Tab from list with pane open: got %v, want FocusDetailPane", m.focus)
	assert.True(t, m.pane.IsOpen(), "Tab must not close the pane")
}

// TestModel_Tab_WrapsDetailsToList verifies Tab wraps details → list.
// T-126: Enter opens the pane with focus on the list, so we need an extra
// Tab to put focus on the pane first.
func TestModel_Tab_WrapsDetailsToList(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // list → details
	require.Equalf(t, appshell.FocusDetailPane, m.focus,
		"precondition: Tab from list with pane open should focus details, got %v", m.focus)

	m = key(m, "tab") // details → list (wrap)
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"Tab from details: got %v, want FocusEntryList", m.focus)
	assert.True(t, m.pane.IsOpen(), "Tab must not close the pane")
}

// TestModel_Tab_NoOpSinglePane verifies Tab is a no-op when the detail pane
// is closed (only the list is visible).
func TestModel_Tab_NoOpSinglePane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "tab")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"Tab with only list visible: got %v, want FocusEntryList", m.focus)
}

// TestModel_Tab_InertWhenFilterPanelFocused verifies Tab does not cycle focus
// while the filter panel (an overlay-like surface) is focused.
func TestModel_Tab_InertWhenFilterPanelFocused(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "f") // focus = FocusFilterPanel

	m = key(m, "tab")
	assert.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"Tab while filter panel focused must be inert, got %v", m.focus)
}

// ---------- T-097: Esc priority 3 (list transient clear) ----------

// TestModel_Esc_OnList_NoTransient_NoOp verifies that Esc on the list when
// nothing transient is set is a benign no-op.
func TestModel_Esc_OnList_NoTransient_NoOp(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	cursorBefore := m.list.Cursor()
	m = key(m, "esc")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after Esc on list: got %v, want FocusEntryList", m.focus)
	assert.Equalf(t, cursorBefore, m.list.Cursor(),
		"cursor changed unexpectedly: before=%d after=%d", cursorBefore, m.list.Cursor())
	assert.False(t, m.pane.IsOpen(), "Esc on list must not open pane")
}

// ---------- T-095: click-to-focus on panes ----------

// TestModel_Click_DetailZone_TransfersFocusToDetail verifies that clicking
// inside the detail-pane zone transfers focus to the detail pane when the
// list was previously focused.
func TestModel_Click_DetailZone_TransfersFocusToDetail(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation at width 200, got %v", m.resize.Orientation())

	// Move focus to the list first.
	m = setFocus(m, appshell.FocusEntryList)

	// Click in the detail-pane zone (well past divider+buffer).
	l := m.layout.Layout()
	detailX := l.ListContentWidth() + 4 // past listEnd buffer + divider + detailStart buffer
	m = send(m, tea.MouseMsg{
		X: detailX, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Equalf(t, appshell.FocusDetailPane, m.focus,
		"click in detail zone (x=%d): focus = %v, want FocusDetailPane", detailX, m.focus)
}

// TestModel_Click_ListZone_TransfersFocusToList verifies that clicking inside
// the entry-list zone returns focus to the list when the detail pane was
// previously focused.
func TestModel_Click_ListZone_TransfersFocusToList(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0]) // T-126: openPane does NOT transfer focus
	// Simulate a prior focus transfer (Tab / click on pane) so we can
	// verify that clicking the list zone returns focus to the list.
	m = setFocus(m, appshell.FocusDetailPane)

	// Click well inside the list area.
	m = send(m, tea.MouseMsg{
		X: 10, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"click in list zone: focus = %v, want FocusEntryList", m.focus)
}

// TestModel_Click_DetailZone_PaneClosed_NoFocusTransfer verifies that
// clicking in what would be the detail-pane zone does NOT transfer focus
// when the pane is closed (the click falls in list territory anyway, but
// guard against accidental focus changes).
func TestModel_Click_DetailZone_PaneClosed_NoFocusTransfer(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	// Pane closed → focus stays on list, no detail zone exists.

	m = send(m, tea.MouseMsg{
		X: 50, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"click with pane closed: focus = %v, want FocusEntryList", m.focus)
}

// ---------- T-126 (F-017, F-024): pane open-time focus policy ----------

// T-126 (F-017): Enter opens the pane with focus STAYING on the list so
// the user can keep navigating entries with `j`/`k` and the pane acts as
// a live preview.
func TestModel_OpenPane_KeepsListFocus(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter")
	require.True(t, m.pane.IsOpen(), "pane should be open after Enter")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after Enter: got %v, want FocusEntryList", m.focus)
}

// T-126 (F-017): OpenDetailPaneMsg (double-click path) also leaves focus
// on the list.
func TestModel_OpenPaneViaMsg_KeepsListFocus(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = send(m, entrylist.OpenDetailPaneMsg{Entry: entries[1]})
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after OpenDetailPaneMsg: got %v, want FocusEntryList", m.focus)
}

// T-126 (F-024): Esc with the list focused and the pane open closes the
// pane — the user should not need to Tab to the pane just to dismiss it.
func TestModel_EscFromList_WithPaneOpen_ClosesPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane, focus stays on list
	require.Truef(t, m.pane.IsOpen() && m.focus == appshell.FocusEntryList,
		"precondition: pane open + list focused; got open=%v focus=%v",
		m.pane.IsOpen(), m.focus)

	m = key(m, "esc")
	assert.False(t, m.pane.IsOpen(), "pane should close on Esc from list-focus")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after Esc: got %v, want FocusEntryList", m.focus)
}

// T-126 (F-017): with the pane open and the list focused, pressing `j`
// moves the list cursor and the detail pane re-renders with the new
// entry (live preview flow).
func TestModel_J_FromList_WithPaneOpen_ReRendersPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = key(m, "enter") // open pane (focus stays on list)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	require.NotNil(t, cmd, "expected SelectionMsg cmd after `j` from list with pane open")
	selMsg := cmd()
	_, ok := selMsg.(entrylist.SelectionMsg)
	require.Truef(t, ok, "expected entrylist.SelectionMsg, got %T", selMsg)
	// Deliver the SelectionMsg to the parent so the pane re-renders.
	m = send(m, selMsg)
	assert.True(t, m.pane.IsOpen(), "pane should still be open after cursor move")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after `j`: got %v, want FocusEntryList (should not transfer)", m.focus)
}

// ---------- T-147 (cavekit-entry-list R13 AC 7): f-focus-transfer clears list search ----------

// TestModel_F_FocusTransfer_ClearsActiveListSearch verifies that pressing
// `f` to open the filter panel while a list search is in navigate mode
// deactivates the search so a stale query + highlights do not linger
// after the user switches contexts (F-108 regression fix). Input-mode
// `f` is a literal query character and is covered by the input-mode path.
func TestModel_F_FocusTransfer_ClearsActiveListSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/")
	m = key(m, "t")
	m = key(m, "enter") // commit to navigate mode so `f` is not captured as query
	require.True(t, m.list.HasActiveSearch(), "precondition: list search should be active")
	require.False(t, m.list.Search().InputMode(),
		"precondition: search should be in navigate mode after Enter")

	m = key(m, "f")

	assert.Equalf(t, appshell.FocusFilterPanel, m.focus,
		"focus after f: got %v, want FocusFilterPanel", m.focus)
	assert.False(t, m.list.HasActiveSearch(), "f focus-transfer should deactivate list search")
	assert.Equalf(t, "", m.list.Search().Query(),
		"list search query should be cleared, got %q", m.list.Search().Query())
}

// ---------- T-153 (cavekit-app-shell R14 AC 5): help overlay Esc preserves list search ----------

// TestModel_HelpOverlay_PreservesListSearchState verifies that opening the
// help overlay (`?`) and dismissing it (`Esc`) over an active, input-mode
// list search leaves the search state fully intact: still active, same
// partial query, still in input mode. The help overlay is a separate
// model from list.search so preservation is by construction — this test
// pins that no future refactor accidentally dismisses the search on
// overlay exit (F-118).
func TestModel_HelpOverlay_PreservesListSearchState(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	m = key(m, "/")
	m = key(m, "a")
	m = key(m, "b")
	m = key(m, "c")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search should be active")
	require.True(t, m.list.Search().InputMode(), "precondition: search should be in input mode")
	require.Equalf(t, "abc", m.list.Search().Query(),
		"precondition: query should be %q, got %q", "abc", m.list.Search().Query())

	m = key(m, "?")
	require.True(t, m.help.IsOpen(), "? should open the help overlay")
	m = key(m, "esc")
	require.False(t, m.help.IsOpen(), "esc should dismiss the help overlay")

	assert.True(t, m.list.HasActiveSearch(), "list search should still be active after help-overlay cycle")
	assert.True(t, m.list.Search().InputMode(), "list search should still be in input mode after help-overlay cycle")
	assert.Equalf(t, "abc", m.list.Search().Query(),
		"list search query should be preserved as %q, got %q", "abc", m.list.Search().Query())
}

// ---------- T-154 (cavekit-entry-list R13 AC 7): mouse-click-off-list clears search ----------

// TestModel_MouseClick_DetailZone_ClearsActiveListSearch verifies that a
// left-button press inside the detail-pane zone while the list has an
// active search both transfers focus AND deactivates the search — the
// paired semantics from T-095 + T-143. The focus-only branch is pinned
// by TestModel_Click_DetailZone_TransfersFocusToDetail; this test pins
// the search-clear branch so the two cannot drift apart (F-119).
func TestModel_MouseClick_DetailZone_ClearsActiveListSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: want right orientation at width 200, got %v", m.resize.Orientation())
	m = setFocus(m, appshell.FocusEntryList)

	m = key(m, "/")
	m = key(m, "t")
	require.True(t, m.list.HasActiveSearch(), "precondition: list search should be active")

	l := m.layout.Layout()
	detailX := l.ListContentWidth() + 4 // past listEnd buffer + divider + detailStart buffer
	m = send(m, tea.MouseMsg{
		X: detailX, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	assert.Equalf(t, appshell.FocusDetailPane, m.focus,
		"click in detail zone: focus = %v, want FocusDetailPane", m.focus)
	assert.False(t, m.list.HasActiveSearch(), "click in detail zone should deactivate list search")
	assert.Equalf(t, "", m.list.Search().Query(),
		"list search query should be cleared, got %q", m.list.Search().Query())
}
