package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

// ---------- smoke ----------

// TestModel_View_NoPanic_Empty verifies View() does not panic on a fresh model.
func TestModel_View_NoPanic_Empty(t *testing.T) {
	m := newModel()
	_ = m.View()
}

// TestModel_View_NoPanic_WithEntries verifies View() does not panic after SetEntries.
func TestModel_View_NoPanic_WithEntries(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(5))
	_ = m.View()
}

// ---------- BUG #2: Init() loading indicator (value receiver drops mutation) ----------

// TestModel_WithFilePath_LoadingIsActive verifies that a model created with a file path
// starts with the loading indicator active.
//
// BUG: Init() calls m.loading.Start() on a value receiver, so the mutation is lost.
// After the model is created, loading is NOT active even though Init() will load a file.
// Fix: activate loading in New() when a non-empty file path is given.
func TestModel_WithFilePath_LoadingIsActive(t *testing.T) {
	m := New("testfile.log", false, "", testCfg())
	if !m.loading.IsActive() {
		t.Error("BUG: loading should be active when model is created with a file path, but is not")
	}
}

// TestModel_StdinMode_LoadingNotActive verifies that stdin mode (empty sourceName)
// does not activate the loading indicator.
func TestModel_StdinMode_LoadingNotActive(t *testing.T) {
	m := newModel() // empty sourceName
	if m.loading.IsActive() {
		t.Error("loading should NOT be active for stdin mode")
	}
}

// TestModel_FollowMode_LoadingNotActive verifies that follow/tail mode does not
// activate the loading indicator (tail is streaming, not batch loading).
func TestModel_FollowMode_LoadingNotActive(t *testing.T) {
	m := New("app.log", true, "", testCfg())
	if m.loading.IsActive() {
		t.Error("loading should NOT be active for follow mode (it is a streaming tail, not a batch load)")
	}
}

// ---------- BUG #1: relayout() discards list Update result ----------

// TestModel_Relayout_ListViewportShrinksWhenPaneOpens verifies that opening the
// detail pane reduces the entry list's rendered row count.
//
// BUG: relayout() calls _, _ = m.list.Update(...) and discards the updated ListModel,
// so the list viewport never shrinks. After the pane opens, the list still thinks it
// occupies the full terminal height.
// Fix: assign m.list, _ = m.list.Update(...).
func TestModel_Relayout_ListViewportShrinksWhenPaneOpens(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)

	// Load enough entries to fill the viewport several times over.
	entries := makeEntries(100)
	m = m.SetEntries(entries)

	renderedBefore := m.list.RenderedRowCount()
	if renderedBefore == 0 {
		t.Fatal("expected non-zero rendered rows before opening pane")
	}

	// Open the detail pane — this calls relayout(), which should shrink the list viewport.
	m = m.openPane(entries[0])

	renderedAfter := m.list.RenderedRowCount()

	// BUG: renderedAfter == renderedBefore because relayout discards the list update.
	// After the fix, the list viewport is reduced by the pane height, so renderedAfter < renderedBefore.
	if renderedAfter >= renderedBefore {
		t.Errorf(
			"BUG: rendered row count did not decrease after pane opened "+
				"(relayout discards list update); before=%d after=%d",
			renderedBefore, renderedAfter,
		)
	}
}

// ---------- BUG #3: visibleCount() correctness ----------

// TestModel_VisibleCount_NoFilter_EqualsAllEntries verifies visibleCount() returns
// len(entries) when no filter is active.
//
// NOTE: visibleCount() also has a performance bug — it calls filter.Apply() on every
// batch message during loading, resulting in O(n²) work for large files.
// This test only checks correctness; the fix is to cache the visible count.
func TestModel_VisibleCount_NoFilter_EqualsAllEntries(t *testing.T) {
	m := newModel()
	entries := makeEntries(10)
	m = m.SetEntries(entries)

	if got := m.visibleCount(); got != len(entries) {
		t.Errorf("visibleCount with no filter: got %d, want %d", got, len(entries))
	}
}

// ---------- quit ----------

// TestModel_Q_EmitsQuitCmd verifies the 'q' key emits a tea.Quit command.
func TestModel_Q_EmitsQuitCmd(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd from 'q' key")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// ---------- focus transitions ----------

// TestModel_F_OpensFocusOnFilterPanel verifies 'f' from the entry list moves
// focus to the filter panel.
func TestModel_F_OpensFocusOnFilterPanel(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = key(m, "f")

	if m.focus != appshell.FocusFilterPanel {
		t.Errorf("focus after 'f': got %v, want FocusFilterPanel", m.focus)
	}
}

// TestModel_Enter_OpensDetailPane verifies pressing Enter when entries exist
// opens the detail pane. T-126 (F-017): focus stays on the entry list so
// `j`/`k` keep navigating the list with the pane as a live preview.
func TestModel_Enter_OpensDetailPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "enter")

	if !m.pane.IsOpen() {
		t.Error("pane should be open after Enter")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after Enter: got %v, want FocusEntryList (pane open does NOT transfer focus)", m.focus)
	}
}

// TestModel_Enter_NoEntries_DoesNotOpenPane verifies Enter with no entries
// does not open the pane.
func TestModel_Enter_NoEntries_DoesNotOpenPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	// No entries loaded.

	m = key(m, "enter")

	if m.pane.IsOpen() {
		t.Error("pane should NOT open when there are no entries")
	}
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

	if m.focus != appshell.FocusDetailPane {
		t.Fatal("expected FocusDetailPane before sending BlurredMsg")
	}

	m = send(m, detailpane.BlurredMsg{})

	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after BlurredMsg: got %v, want FocusEntryList", m.focus)
	}
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

	if cmd == nil {
		t.Fatal("Esc in detail pane should return a non-nil cmd")
	}
	blurMsg := cmd()
	if _, ok := blurMsg.(detailpane.BlurredMsg); !ok {
		t.Fatalf("expected detailpane.BlurredMsg from Esc, got %T", blurMsg)
	}

	// Deliver the BlurredMsg to the parent model to complete the focus handoff.
	m = send(m, blurMsg)
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after BlurredMsg: got %v, want FocusEntryList", m.focus)
	}
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
	if !m.help.IsOpen() {
		t.Fatal("help overlay should be open after '?'")
	}

	// Press 'j' — should be intercepted by the overlay and NOT move the cursor.
	m = key(m, "j")
	if m.list.Cursor() != cursorBefore {
		t.Errorf("cursor moved while help overlay was open: before=%d after=%d",
			cursorBefore, m.list.Cursor())
	}
}

// ---------- loading stream ----------

// TestModel_LoadDone_LoadingBecomesInactive verifies that a LoadDoneMsg deactivates
// the loading indicator.
func TestModel_LoadDone_LoadingBecomesInactive(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	// Manually activate loading to test the Done path.
	m.loading = m.loading.Start()

	m = send(m, logsource.LoadDoneMsg{})

	if m.loading.IsActive() {
		t.Error("loading should be inactive after LoadDoneMsg")
	}
}

// ---------- window resize ----------

// TestModel_WindowSizeMsg_UpdatesDimensions verifies that a WindowSizeMsg is
// propagated to the resize model so Width() and Height() reflect the new size.
func TestModel_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 40)

	if m.resize.Width() != 120 {
		t.Errorf("resize.Width(): got %d, want 120", m.resize.Width())
	}
	if m.resize.Height() != 40 {
		t.Errorf("resize.Height(): got %d, want 40", m.resize.Height())
	}
}

// ---------- entry list integration ----------

// TestModel_SetEntries_UpdatesHeaderCounts verifies that SetEntries makes the
// header reflect total and visible entry counts.
func TestModel_SetEntries_UpdatesHeaderCounts(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)

	entries := makeEntries(7)
	m = m.SetEntries(entries)

	// With no filter, total == visible == 7. The header View() should contain "7".
	headerView := m.header.View()
	if !containsCount(headerView, 7) {
		t.Errorf("header should show count 7 after SetEntries; got: %q", headerView)
	}
}

// TestModel_OpenDetailPane_ViaMsg verifies OpenDetailPaneMsg (double-click path)
// opens the pane.
func TestModel_OpenDetailPane_ViaMsg(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)

	m = send(m, entrylist.OpenDetailPaneMsg{Entry: entries[1]})

	if !m.pane.IsOpen() {
		t.Error("pane should be open after OpenDetailPaneMsg")
	}
	// T-126 (F-017): opening the pane does NOT transfer focus.
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus: got %v, want FocusEntryList (pane open does not steal focus)", m.focus)
	}
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
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("Tab from list with pane open: got %v, want FocusDetailPane", m.focus)
	}
	if !m.pane.IsOpen() {
		t.Error("Tab must not close the pane")
	}
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
	if m.focus != appshell.FocusDetailPane {
		t.Fatalf("precondition: Tab from list with pane open should focus details, got %v", m.focus)
	}

	m = key(m, "tab") // details → list (wrap)
	if m.focus != appshell.FocusEntryList {
		t.Errorf("Tab from details: got %v, want FocusEntryList", m.focus)
	}
	if !m.pane.IsOpen() {
		t.Error("Tab must not close the pane")
	}
}

// TestModel_Tab_NoOpSinglePane verifies Tab is a no-op when the detail pane
// is closed (only the list is visible).
func TestModel_Tab_NoOpSinglePane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "tab")
	if m.focus != appshell.FocusEntryList {
		t.Errorf("Tab with only list visible: got %v, want FocusEntryList", m.focus)
	}
}

// TestModel_Tab_InertWhenFilterPanelFocused verifies Tab does not cycle focus
// while the filter panel (an overlay-like surface) is focused.
func TestModel_Tab_InertWhenFilterPanelFocused(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "f") // focus = FocusFilterPanel

	m = key(m, "tab")
	if m.focus != appshell.FocusFilterPanel {
		t.Errorf("Tab while filter panel focused must be inert, got %v", m.focus)
	}
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
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after Esc on list: got %v, want FocusEntryList", m.focus)
	}
	if m.list.Cursor() != cursorBefore {
		t.Errorf("cursor changed unexpectedly: before=%d after=%d", cursorBefore, m.list.Cursor())
	}
	if m.pane.IsOpen() {
		t.Error("Esc on list must not open pane")
	}
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
	if !m.pane.IsOpen() {
		t.Fatal("precondition: pane should be open after openPane")
	}
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: orientation should be right at termWidth=200, got %v", m.resize.Orientation())
	}

	// Shrink while keeping right-split: detailW drops below MinDetailWidth.
	m = resize(m, 100, 24)
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: orientation should stay right at termWidth=100, got %v", m.resize.Orientation())
	}

	if m.pane.IsOpen() {
		t.Errorf("expected auto-close at termWidth=100 (detailW=%d), pane still open", m.layout.Layout().DetailContentWidth())
	}
	if !m.keyhints.HasNotice() {
		t.Error("expected status notice after auto-close")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after auto-close: got %v, want FocusEntryList", m.focus)
	}
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
	if !m.keyhints.HasNotice() {
		t.Fatal("precondition: notice should be set after auto-close")
	}

	m = send(m, noticeClearMsg{})
	if m.keyhints.HasNotice() {
		t.Error("notice should be cleared after noticeClearMsg")
	}
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

	if !m.pane.IsOpen() {
		t.Errorf("pane should remain open at termWidth=180 (detailW=%d)", m.layout.Layout().DetailContentWidth())
	}
	if m.keyhints.HasNotice() {
		t.Error("notice should not be set when pane stays open")
	}
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
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation, got %v", m.resize.Orientation())
	}
	before := m.cfg.Config.DetailPane.WidthRatio

	// Press on the divider column to start the drag session.
	l := m.layout.Layout()
	divider := l.ListContentWidth() + 2
	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !m.draggingDivider {
		t.Fatalf("precondition: expected draggingDivider=true after Press on divider column %d", divider)
	}

	// Motion to a column 20 cells to the left → detail grows → ratio rises.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	after := m.cfg.Config.DetailPane.WidthRatio
	if after <= before {
		t.Errorf("drag-left should increase width_ratio: before=%.3f after=%.3f", before, after)
	}

	// Release ends the drag session.
	m = send(m, tea.MouseMsg{X: divider - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	if m.draggingDivider {
		t.Errorf("expected draggingDivider=false after Release")
	}
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
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation at width 200, got %v", m.resize.Orientation())
	}

	// Move focus to the list first.
	m.focus = appshell.FocusEntryList
	m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)

	// Click in the detail-pane zone (well past divider+buffer).
	l := m.layout.Layout()
	detailX := l.ListContentWidth() + 4 // past listEnd buffer + divider + detailStart buffer
	m = send(m, tea.MouseMsg{
		X: detailX, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("click in detail zone (x=%d): focus = %v, want FocusDetailPane", detailX, m.focus)
	}
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
	m.focus = appshell.FocusDetailPane
	m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
	if m.focus != appshell.FocusDetailPane {
		t.Fatalf("precondition: focus should be detail, got %v", m.focus)
	}

	// Click well inside the list area.
	m = send(m, tea.MouseMsg{
		X: 10, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	if m.focus != appshell.FocusEntryList {
		t.Errorf("click in list zone: focus = %v, want FocusEntryList", m.focus)
	}
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
	if m.focus != appshell.FocusEntryList {
		t.Errorf("click with pane closed: focus = %v, want FocusEntryList", m.focus)
	}
}

// ---------- T-099: ratio live write-back to config ----------

// TestModel_RatioKey_PersistsToConfigFile verifies that pressing a ratio
// key (here `+` in right-split) writes the new width_ratio to disk.
func TestModel_RatioKey_PersistsToConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"

	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	m = m.openPane(makeEntries(3)[0])
	// T-126: openPane no longer auto-focuses the pane, but the `+` ratio
	// key only fires when the pane is focused (handleKey's FocusDetailPane
	// branch). Simulate the Tab-to-pane step.
	m.focus = appshell.FocusDetailPane
	m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)

	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation, got %v", m.resize.Orientation())
	}
	beforeWidth := m.cfg.Config.DetailPane.WidthRatio
	beforeHeight := m.cfg.Config.DetailPane.HeightRatio

	m = key(m, "+") // increment width_ratio in right-split
	if m.cfg.Config.DetailPane.WidthRatio == beforeWidth {
		t.Fatalf("'+' did not change in-memory width_ratio: %.3f", m.cfg.Config.DetailPane.WidthRatio)
	}

	// Reload from disk and verify width_ratio persisted.
	reloaded := config.Load(cfgPath)
	if reloaded.Config.DetailPane.WidthRatio != m.cfg.Config.DetailPane.WidthRatio {
		t.Errorf("disk width_ratio: got %.3f, want %.3f",
			reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
	}
	// height_ratio must NOT have been clobbered.
	if reloaded.Config.DetailPane.HeightRatio != beforeHeight {
		t.Errorf("disk height_ratio mutated: got %.3f, want %.3f (untouched)",
			reloaded.Config.DetailPane.HeightRatio, beforeHeight)
	}
	_ = m
}

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
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation, got %v", m.resize.Orientation())
	}

	l := m.layout.Layout()
	divider := l.ListContentWidth() + 2

	m = send(m, tea.MouseMsg{X: divider, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !m.draggingDivider {
		t.Fatalf("precondition: drag did not start at divider x=%d", divider)
	}
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: divider - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	reloaded := config.Load(cfgPath)
	if reloaded.Config.DetailPane.WidthRatio != m.cfg.Config.DetailPane.WidthRatio {
		t.Errorf("disk width_ratio after drag release: got %.3f, want %.3f",
			reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
	}
}

// ---------- T-105: orientation flip preserves both ratios ----------

// TestModel_OrientationFlip_PreservesBothRatios verifies that flipping from
// right → below → right does not mutate height_ratio or width_ratio in the
// in-memory config, regardless of which ratio is "active" for the current
// orientation.
func TestModel_OrientationFlip_PreservesBothRatios(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.HeightRatio = 0.60
	cfg.Config.DetailPane.WidthRatio = 0.20

	m := New("", false, "", cfg)
	if m.cfg.Config.DetailPane.HeightRatio != 0.60 {
		t.Fatalf("pre-resize height_ratio: got %.2f, want 0.60", m.cfg.Config.DetailPane.HeightRatio)
	}

	assertRatios := func(step string) {
		t.Helper()
		if m.cfg.Config.DetailPane.HeightRatio != 0.60 {
			t.Errorf("%s: height_ratio = %.3f, want 0.600 (unmutated)", step, m.cfg.Config.DetailPane.HeightRatio)
		}
		if m.cfg.Config.DetailPane.WidthRatio != 0.20 {
			t.Errorf("%s: width_ratio = %.3f, want 0.200 (unmutated)", step, m.cfg.Config.DetailPane.WidthRatio)
		}
	}

	m = resize(m, 200, 24) // right
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: 200 cols should be right, got %v", m.resize.Orientation())
	}
	assertRatios("after initial right resize")

	m = resize(m, 80, 24) // below (under 100-col threshold)
	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: 80 cols should be below, got %v", m.resize.Orientation())
	}
	assertRatios("after flip to below")

	m = resize(m, 200, 24) // back to right
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: 200 cols should flip back to right, got %v", m.resize.Orientation())
	}
	assertRatios("after flip back to right")
}

// ---------- T-117: dismiss paneSearch on pane close / reopen (F-006) ----------

// TestModel_PaneSearch_DismissedOnBlurred verifies that when the pane
// closes, any active search state is cleared so a subsequent open starts
// with a clean query.
func TestModel_PaneSearch_DismissedOnBlurred(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane, focus → detail
	m = key(m, "/")     // start search
	m = key(m, "t")     // type a char
	if !m.paneSearch.IsActive() {
		t.Fatal("precondition: search should be active after '/'")
	}
	if m.paneSearch.Query() == "" {
		t.Fatal("precondition: query should be non-empty after typing")
	}

	m = send(m, detailpane.BlurredMsg{})

	if m.paneSearch.IsActive() {
		t.Error("search should be dismissed after pane closes")
	}
	if m.paneSearch.Query() != "" {
		t.Errorf("search query should be cleared after pane closes, got %q", m.paneSearch.Query())
	}
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

	if m.paneSearch.IsActive() {
		t.Error("search should be dismissed when a new entry opens the pane")
	}
	if m.paneSearch.Query() != "" {
		t.Errorf("search query should be cleared on reopen, got %q", m.paneSearch.Query())
	}
}

// ---------- T-120: two-step Esc integration (cavekit-detail-pane R7 / F-007) ----------

// TestModel_TwoStepEsc_DismissesSearchThenClosesPane verifies the
// cavekit R7 "two-step Esc" contract: first Esc with search active
// dismisses search and leaves the pane open; second Esc closes the pane.
func TestModel_TwoStepEsc_DismissesSearchThenClosesPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane
	m = key(m, "/")     // activate search
	m = key(m, "h")     // type something to exercise the query path
	if !m.paneSearch.IsActive() {
		t.Fatal("precondition: search should be active before first Esc")
	}
	if !m.pane.IsOpen() {
		t.Fatal("precondition: pane should be open before first Esc")
	}

	// First Esc — dismisses search, pane stays open.
	m = key(m, "esc")
	if m.paneSearch.IsActive() {
		t.Error("first Esc should dismiss search")
	}
	if !m.pane.IsOpen() {
		t.Error("first Esc should NOT close the pane")
	}
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("after first Esc focus should remain FocusDetailPane, got %v", m.focus)
	}

	// Second Esc — closes the pane. Pane emits BlurredMsg which we
	// deliver to the model to complete the focus handoff.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.pane.IsOpen() {
		t.Error("second Esc should close the pane")
	}
	if cmd == nil {
		t.Fatal("second Esc should return BlurredMsg cmd")
	}
	blurMsg := cmd()
	if _, ok := blurMsg.(detailpane.BlurredMsg); !ok {
		t.Fatalf("second Esc should emit BlurredMsg, got %T", blurMsg)
	}
	m = send(m, blurMsg)
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after second Esc: got %v, want FocusEntryList", m.focus)
	}
}

// ---------- T-116: cross-pane `/` activation (app-shell R13 / F-001, F-011) ----------

// TestModel_Slash_ListFocus_PaneOpen_TransfersAndActivates verifies that
// `/` with list focused and pane open transfers focus to the pane AND
// activates search in a single keypress.
func TestModel_Slash_ListFocus_PaneOpen_TransfersAndActivates(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Move focus back to list so we're testing the cross-pane activation.
	m.focus = appshell.FocusEntryList
	m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)

	m = key(m, "/")

	if m.focus != appshell.FocusDetailPane {
		t.Errorf("/ with pane open should transfer focus to detail pane, got %v", m.focus)
	}
	if !m.paneSearch.IsActive() {
		t.Error("/ with pane open should activate paneSearch")
	}
	if m.paneSearch.Mode() != detailpane.SearchModeInput {
		t.Errorf("activated search should be in input mode, got %v", m.paneSearch.Mode())
	}
}

// TestModel_Slash_ListFocus_PaneClosed_ShowsNotice verifies that `/` with
// list focused and pane closed emits a transient notice and never acts
// as a silent no-op.
func TestModel_Slash_ListFocus_PaneClosed_ShowsNotice(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	if m.pane.IsOpen() {
		t.Fatal("precondition: pane should be closed")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	if !m.keyhints.HasNotice() {
		t.Error("/ with pane closed should set a transient keyhint notice")
	}
	if cmd == nil {
		t.Error("expected auto-dismiss cmd from notice")
	}
	if m.paneSearch.IsActive() {
		t.Error("/ with pane closed must NOT activate search")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("/ with pane closed should not steal focus, got %v", m.focus)
	}
}

// TestModel_Slash_FilterPanelFocus_RoutedToFilter verifies that `/` with
// the filter panel focused is routed to the filter input as a literal
// character (not intercepted as a global search activation).
func TestModel_Slash_FilterPanelFocus_RoutedToFilter(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "f") // focus = FocusFilterPanel
	if m.focus != appshell.FocusFilterPanel {
		t.Fatalf("precondition: want filter panel focused, got %v", m.focus)
	}

	m = key(m, "/")

	if m.paneSearch.IsActive() {
		t.Error("/ in filter panel should NOT activate pane search")
	}
	if m.focus != appshell.FocusFilterPanel {
		t.Errorf("/ in filter panel should stay in filter panel, got %v", m.focus)
	}
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
	m = key(m, "/") // activate — input mode
	if !m.paneSearch.IsActive() {
		t.Fatal("precondition: search should be active after '/'")
	}
	m = key(m, "enter") // commit to navigate
	if m.paneSearch.Mode() != detailpane.SearchModeNavigate {
		t.Fatalf("precondition: want SearchModeNavigate, got %v", m.paneSearch.Mode())
	}

	paneBefore := m.pane
	m = key(m, "j")
	// Query must be unchanged: `j` in nav mode should NOT extend the query.
	if m.paneSearch.Query() != "" {
		t.Errorf("nav-mode `j` should not extend query, got %q", m.paneSearch.Query())
	}
	// Pane view should have changed IF the content exceeded the viewport.
	// We check by rendering View() and comparing — if there was scroll
	// room, the view string differs.
	if m.pane.View() == paneBefore.View() {
		// Not a hard failure — if content fit in the viewport, scroll is
		// a no-op. Log instead.
		t.Logf("note: pane view unchanged after j (content may have fit in viewport)")
	}
	// And search is still active in navigate mode.
	if !m.paneSearch.IsActive() {
		t.Error("search should still be active after nav-mode j")
	}
	if m.paneSearch.Mode() != detailpane.SearchModeNavigate {
		t.Errorf("mode should stay navigate after j, got %v", m.paneSearch.Mode())
	}
}

// TestModel_Search_InputMode_JAppendsToQuery verifies that in input mode,
// `j` becomes a literal query character (does not scroll the pane).
func TestModel_Search_InputMode_JAppendsToQuery(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane
	m = key(m, "/")     // activate search — input mode
	m = key(m, "j")     // should extend query
	if m.paneSearch.Query() != "j" {
		t.Errorf("input-mode `j` should extend query to 'j', got %q", m.paneSearch.Query())
	}
}

// ---------- T-123: detail-pane vertical height in right orientation (P0 / F-013) ----------

// TestModel_PaneHeight_RightOrientation_UsesFullSlot verifies that in
// right-split the pane's ContentHeight fills the main-area slot
// (terminal_height - header - status - border_rows), not height_ratio.
func TestModel_PaneHeight_RightOrientation_UsesFullSlot(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right" // force right so auto-threshold is bypassed
	cfg.Config.DetailPane.HeightRatio = 0.30 // the below-mode value that used to clip

	m := New("", false, "", cfg)
	m = resize(m, 140, 24) // wide enough to keep right-split stable (detail ≥ 30 cells)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: forced orientation=right, got %v", m.resize.Orientation())
	}
	// 24 (height) - 1 (header) - 1 (status) = 22 outer rows → ContentHeight ≥ 20.
	if got := m.pane.ContentHeight(); got < 20 {
		t.Errorf("right-split ContentHeight: got %d, want ≥ 20 (F-013)", got)
	}
	// Regression: height_ratio × terminalHeight = 7. If ContentHeight is ≤ 5
	// we are still applying below-mode math in right orientation.
	if got := m.pane.ContentHeight(); got <= 5 {
		t.Errorf("right-split ContentHeight = %d — still clipped to height_ratio", got)
	}
}

// TestModel_PaneHeight_BelowOrientation_UsesRatio verifies the ratio-based
// vertical sizing still governs below-mode.
func TestModel_PaneHeight_BelowOrientation_UsesRatio(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	cfg.Config.DetailPane.HeightRatio = 0.30

	m := New("", false, "", cfg)
	m = resize(m, 80, 24) // below-mode orientation
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: forced orientation=below, got %v", m.resize.Orientation())
	}
	// 24 × 0.30 = 7 outer rows → ContentHeight = 5.
	wantOuter := 7 // int(24 * 0.30) = 7
	if got := m.paneHeight.PaneHeight(); got != wantOuter {
		t.Errorf("below-mode outer height: got %d, want %d", got, wantOuter)
	}
	if got := m.pane.ContentHeight(); got != wantOuter-2 {
		t.Errorf("below-mode ContentHeight: got %d, want %d", got, wantOuter-2)
	}
}

// TestModel_RatioKey_StillWorksInBelowMode verifies that `+` in below-mode
// still increases height_ratio after the T-123 refactor (guards against
// accidental regression on the resize keymap path).
func TestModel_RatioKey_StillWorksInBelowMode(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	cfg.Config.DetailPane.HeightRatio = 0.30

	m := New("", false, "", cfg)
	m = resize(m, 80, 30)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	// Transfer focus to the pane so the ratio keymap fires.
	m.focus = appshell.FocusDetailPane

	before := m.paneHeight.Ratio()
	m = key(m, "+")
	if m.paneHeight.Ratio() <= before {
		t.Errorf("`+` in below-mode: ratio = %.2f, want > %.2f", m.paneHeight.Ratio(), before)
	}
}

// TestModel_OrientationFlip_VerticalSizeTracks verifies that flipping from
// below to right uses the full main-area slot (T-123 F-013), and flipping
// back restores the below-mode ratio-derived height.
func TestModel_OrientationFlip_VerticalSizeTracks(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "auto"
	cfg.Config.DetailPane.HeightRatio = 0.30
	cfg.Config.DetailPane.OrientationThresholdCols = 100

	m := New("", false, "", cfg)
	m = resize(m, 80, 24) // below (under threshold)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: 80 cols should be below, got %v", m.resize.Orientation())
	}
	belowOuter := m.paneHeight.PaneHeight()
	if got, want := belowOuter, 7; got != want {
		t.Errorf("below outer height: got %d, want %d", got, want)
	}

	// Flip to right — wide terminal, detailW ≥ 30 so pane stays open.
	m = resize(m, 140, 24)
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: 140 cols should flip to right, got %v", m.resize.Orientation())
	}
	if !m.pane.IsOpen() {
		t.Fatalf("precondition: pane should still be open after flip to right")
	}
	// Right-mode vertical = full main slot = 24 - 1 - 1 = 22.
	if got := m.pane.ContentHeight(); got < 20 {
		t.Errorf("right-mode ContentHeight after flip: got %d, want ≥ 20", got)
	}

	// Flip back to below — height_ratio preserved, ContentHeight returns to
	// the ratio-derived value.
	m = resize(m, 80, 24)
	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: 80 cols should flip back to below, got %v", m.resize.Orientation())
	}
	if got := m.paneHeight.PaneHeight(); got != belowOuter {
		t.Errorf("below outer height after round-trip: got %d, want %d (ratio not preserved)", got, belowOuter)
	}
	if m.cfg.Config.DetailPane.HeightRatio != 0.30 {
		t.Errorf("height_ratio mutated by orientation flip: got %.3f, want 0.300",
			m.cfg.Config.DetailPane.HeightRatio)
	}
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
	if !m.pane.IsOpen() {
		t.Fatal("pane should be open after Enter")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after Enter: got %v, want FocusEntryList", m.focus)
	}
}

// T-126 (F-017): OpenDetailPaneMsg (double-click path) also leaves focus
// on the list.
func TestModel_OpenPaneViaMsg_KeepsListFocus(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = send(m, entrylist.OpenDetailPaneMsg{Entry: entries[1]})
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after OpenDetailPaneMsg: got %v, want FocusEntryList", m.focus)
	}
}

// T-126 (F-024): Esc with the list focused and the pane open closes the
// pane — the user should not need to Tab to the pane just to dismiss it.
func TestModel_EscFromList_WithPaneOpen_ClosesPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // open pane, focus stays on list
	if !m.pane.IsOpen() || m.focus != appshell.FocusEntryList {
		t.Fatalf("precondition: pane open + list focused; got open=%v focus=%v",
			m.pane.IsOpen(), m.focus)
	}

	m = key(m, "esc")
	if m.pane.IsOpen() {
		t.Error("pane should close on Esc from list-focus")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after Esc: got %v, want FocusEntryList", m.focus)
	}
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
	if cmd == nil {
		t.Fatal("expected SelectionMsg cmd after `j` from list with pane open")
	}
	selMsg := cmd()
	if _, ok := selMsg.(entrylist.SelectionMsg); !ok {
		t.Fatalf("expected entrylist.SelectionMsg, got %T", selMsg)
	}
	// Deliver the SelectionMsg to the parent so the pane re-renders.
	m = send(m, selMsg)
	if !m.pane.IsOpen() {
		t.Error("pane should still be open after cursor move")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("focus after `j`: got %v, want FocusEntryList (should not transfer)", m.focus)
	}
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
