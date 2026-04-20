package app

import (
	"os"
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
	// Post-T-160 the visible `│` column equals ListContentWidth().
	l := m.layout.Layout()
	divider := l.ListContentWidth()
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
	divider := l.ListContentWidth() // post-T-160 visible-glyph column

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
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144: `/` is focus-based)
	m = key(m, "/")     // start pane search
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
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144: focus-based `/`)
	m = key(m, "/")     // activate pane search
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
	if m.focus != appshell.FocusEntryList {
		t.Fatalf("precondition: want list focus after openPane, got %v", m.focus)
	}

	m = key(m, "/")

	if m.focus != appshell.FocusEntryList {
		t.Errorf("/ with list focused must NOT transfer focus, got %v", m.focus)
	}
	if !m.list.HasActiveSearch() {
		t.Error("/ with list focused should activate list search")
	}
	if m.paneSearch.IsActive() {
		t.Error("/ with list focused must NOT activate pane search")
	}
}

// TestModel_Slash_ListFocus_PaneClosed_ActivatesListSearch verifies that
// `/` with list focused and pane closed activates list search (T-144).
// The old T-116 "open entry first" notice is gone — list search is
// always available when the list is focused.
func TestModel_Slash_ListFocus_PaneClosed_ActivatesListSearch(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	if m.pane.IsOpen() {
		t.Fatal("precondition: pane should be closed")
	}

	m = key(m, "/")

	if !m.list.HasActiveSearch() {
		t.Error("/ with list focused (pane closed) should activate list search")
	}
	if m.paneSearch.IsActive() {
		t.Error("/ with list focused must NOT activate pane search")
	}
	if m.focus != appshell.FocusEntryList {
		t.Errorf("/ with list focused should stay on list, got %v", m.focus)
	}
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

	if !m.paneSearch.IsActive() {
		t.Error("/ with pane focused should activate pane search")
	}
	if m.list.HasActiveSearch() {
		t.Error("/ with pane focused must NOT activate list search")
	}
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
	if !m.list.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}

	m = key(m, "tab") // focus → detail pane

	if m.list.HasActiveSearch() {
		t.Error("Tab cycling off the list should clear list search")
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
	m = key(m, "tab") // focus detail pane (T-144: `/` is focus-based)
	m = key(m, "/")   // activate — input mode
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
	m = key(m, "enter") // open pane (focus stays on list per T-126)
	m = key(m, "tab")   // Tab → focus detail pane (T-144)
	m = key(m, "/")     // activate pane search — input mode
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
	if m.list.Cursor() != 2 {
		t.Fatalf("precondition: cursor at last entry, got %d", m.list.Cursor())
	}
	m = key(m, "enter")
	if !m.pane.IsOpen() {
		t.Fatal("precondition: pane open after Enter")
	}
	if !containsSubstring(m.pane.View(), "gamma-msg") {
		t.Fatalf("precondition: pane shows gamma-msg; got: %q", m.pane.View())
	}

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entry: appended}))

	if m.list.Cursor() != 3 {
		t.Errorf("cursor should snap to new last entry (3), got %d", m.list.Cursor())
	}
	if !m.pane.IsOpen() {
		t.Error("pane should still be open after append")
	}
	post := m.pane.View()
	if !containsSubstring(post, "delta-unique") {
		t.Errorf("pane must re-render with appended entry (delta-unique); got: %q", post)
	}
	if containsSubstring(post, "gamma-msg") {
		t.Errorf("pane must not keep rendering previous entry (gamma-msg); got: %q", post)
	}
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
	if !m.pane.IsOpen() || m.list.Cursor() != 1 {
		t.Fatalf("precondition: cursor=%d pane=%v", m.list.Cursor(), m.pane.IsOpen())
	}

	batch := logsource.EntryBatchMsg{Entries: []logsource.Entry{
		jsonEntry(3, "charlie-msg"),
		jsonEntry(4, "delta-unique"),
	}}
	m = send(m, logsource.NewLoadFileStreamMsgForTest(batch))

	if m.list.Cursor() != 3 {
		t.Errorf("cursor should snap to new last entry (3), got %d", m.list.Cursor())
	}
	post := m.pane.View()
	if !containsSubstring(post, "delta-unique") {
		t.Errorf("pane must re-render with last appended entry (delta-unique); got: %q", post)
	}
	if containsSubstring(post, "beta-msg") {
		t.Errorf("pane must not keep rendering pre-append entry (beta-msg); got: %q", post)
	}
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
	if !m.pane.IsOpen() || !containsSubstring(m.pane.View(), "alpha-msg") {
		t.Fatalf("precondition: pane open on alpha; got: %q", m.pane.View())
	}

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entry: appended}))

	if m.list.Cursor() != 0 {
		t.Errorf("cursor must NOT move when pre-append cursor < tail, got %d", m.list.Cursor())
	}
	post := m.pane.View()
	if !containsSubstring(post, "alpha-msg") {
		t.Errorf("pane must still show alpha-msg (user's non-tail selection preserved); got: %q", post)
	}
	if containsSubstring(post, "delta-unique") {
		t.Errorf("pane must NOT render appended entry (delta-unique) when follow disengaged; got: %q", post)
	}
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
	if containsSubstring(view, "secret") {
		t.Errorf("detail pane view should not include suppressed field `secret`, got: %q", view)
	}
	if containsSubstring(view, "shh") {
		t.Errorf("detail pane view should not include suppressed value, got: %q", view)
	}
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
	if !m.list.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}
	if !m.list.Search().InputMode() {
		t.Fatal("precondition: list search should be in input mode")
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)

	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, isQuit := msg.(tea.QuitMsg); isQuit {
				t.Fatal("q in list-search input mode must NOT quit")
			}
		}
	}
	if m.list.Search().Query() != "q" {
		t.Errorf("q in input mode should extend query to 'q', got %q", m.list.Search().Query())
	}
	if !m.list.HasActiveSearch() {
		t.Error("list search should still be active after typed q")
	}

	// F-117: chained keystrokes — the whole word "quit" must land in the query
	// without the trailing "uit" ever escaping through to the global quit
	// handler. Assert no tea.QuitMsg across the full sequence.
	for _, r := range []rune{'u', 'i', 't'} {
		updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
		if cmd != nil {
			if msg := cmd(); msg != nil {
				if _, isQuit := msg.(tea.QuitMsg); isQuit {
					t.Fatalf("keystroke %q in list-search input mode must NOT quit", string(r))
				}
			}
		}
	}
	if got := m.list.Search().Query(); got != "quit" {
		t.Errorf("chained q/u/i/t should build query %q, got %q", "quit", got)
	}
	if !m.list.HasActiveSearch() {
		t.Error("list search should still be active after chained keystrokes")
	}
	if !m.list.Search().InputMode() {
		t.Error("search should still be in input mode after chained keystrokes")
	}
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
	m = key(m, "t") // query "t" matches "test message"
	m = key(m, "enter") // commit → navigate mode
	if m.list.Search().InputMode() {
		t.Fatal("precondition: search should be in navigate mode after Enter")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q in navigate mode should still emit quit cmd")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("q in navigate mode should emit tea.QuitMsg")
	}
}

// TestModel_Q_NoListSearch_StillQuits verifies the baseline quit behaviour
// is preserved when no list search is active.
func TestModel_Q_NoListSearch_StillQuits(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q without active search should quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected tea.QuitMsg")
	}
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
	if m.focus != appshell.FocusDetailPane {
		t.Fatalf("precondition: want detail pane focus, got %v", m.focus)
	}

	// `q` on detail pane focus is NOT a quit (global quit only triggers on
	// list focus — this is unchanged by T-146). Just confirms no regression.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, isQuit := msg.(tea.QuitMsg); isQuit {
				t.Error("q on detail pane focus should NOT quit (global quit requires list focus)")
			}
		}
	}
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
	if !m.list.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}
	if m.list.Search().InputMode() {
		t.Fatal("precondition: search should be in navigate mode after Enter")
	}

	m = key(m, "f")

	if m.focus != appshell.FocusFilterPanel {
		t.Errorf("focus after f: got %v, want FocusFilterPanel", m.focus)
	}
	if m.list.HasActiveSearch() {
		t.Error("f focus-transfer should deactivate list search")
	}
	if m.list.Search().Query() != "" {
		t.Errorf("list search query should be cleared, got %q", m.list.Search().Query())
	}
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
	if !m.list.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}
	if !m.list.Search().InputMode() {
		t.Fatal("precondition: search should be in input mode")
	}
	if m.list.Search().Query() != "abc" {
		t.Fatalf("precondition: query should be %q, got %q", "abc", m.list.Search().Query())
	}

	m = key(m, "?")
	if !m.help.IsOpen() {
		t.Fatal("? should open the help overlay")
	}
	m = key(m, "esc")
	if m.help.IsOpen() {
		t.Fatal("esc should dismiss the help overlay")
	}

	if !m.list.HasActiveSearch() {
		t.Error("list search should still be active after help-overlay cycle")
	}
	if !m.list.Search().InputMode() {
		t.Error("list search should still be in input mode after help-overlay cycle")
	}
	if got := m.list.Search().Query(); got != "abc" {
		t.Errorf("list search query should be preserved as %q, got %q", "abc", got)
	}
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
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation at width 200, got %v", m.resize.Orientation())
	}
	m.focus = appshell.FocusEntryList
	m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)

	m = key(m, "/")
	m = key(m, "t")
	if !m.list.HasActiveSearch() {
		t.Fatal("precondition: list search should be active")
	}

	l := m.layout.Layout()
	detailX := l.ListContentWidth() + 4 // past listEnd buffer + divider + detailStart buffer
	m = send(m, tea.MouseMsg{
		X: detailX, Y: 5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})

	if m.focus != appshell.FocusDetailPane {
		t.Errorf("click in detail zone: focus = %v, want FocusDetailPane", m.focus)
	}
	if m.list.HasActiveSearch() {
		t.Error("click in detail zone should deactivate list search")
	}
	if got := m.list.Search().Query(); got != "" {
		t.Errorf("list search query should be cleared, got %q", got)
	}
}

// ---------- T-155: focus-aware keyboard resize (cavekit-app-shell R12 revised) ----------

// setupRatioModelRight returns a model in right-orientation with the
// detail pane open. Ratios are the defaults: detail width_ratio = 0.30,
// so list share = 0.70.
func setupRatioModelRight(t *testing.T, listFocus bool) Model {
	t.Helper()
	dir := t.TempDir()
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, dir+"/config.toml", cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation, got %v", m.resize.Orientation())
	}
	if listFocus {
		m.focus = appshell.FocusEntryList
		m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)
	} else {
		m.focus = appshell.FocusDetailPane
		m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
	}
	return m
}

// setupRatioModelBelow returns a model in below-orientation with the
// detail pane open.
func setupRatioModelBelow(t *testing.T, listFocus bool) Model {
	t.Helper()
	dir := t.TempDir()
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, dir+"/config.toml", cfg)
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: want below orientation, got %v", m.resize.Orientation())
	}
	if listFocus {
		m.focus = appshell.FocusEntryList
		m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)
	} else {
		m.focus = appshell.FocusDetailPane
		m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
	}
	return m
}

// TestModel_T155_Plus_DetailFocus_GrowsDetail_Right: `+` with detail
// focused in right-mode grows the detail width_ratio.
func TestModel_T155_Plus_DetailFocus_GrowsDetail_Right(t *testing.T) {
	m := setupRatioModelRight(t, false)
	before := m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "+")
	after := m.cfg.Config.DetailPane.WidthRatio
	if after <= before {
		t.Errorf("+ detail-focus must grow detail ratio: before=%.3f after=%.3f", before, after)
	}
}

// TestModel_T155_Plus_ListFocus_ShrinksDetail_Right: `+` with list
// focused in right-mode shrinks detail width_ratio (list grows).
func TestModel_T155_Plus_ListFocus_ShrinksDetail_Right(t *testing.T) {
	m := setupRatioModelRight(t, true)
	before := m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "+")
	after := m.cfg.Config.DetailPane.WidthRatio
	if after >= before {
		t.Errorf("+ list-focus must shrink detail ratio: before=%.3f after=%.3f", before, after)
	}
}

// TestModel_T155_Minus_SymmetricInverse_Right: `-` at each focus is the
// inverse of `+`.
func TestModel_T155_Minus_SymmetricInverse_Right(t *testing.T) {
	// detail focus: `-` shrinks detail
	m := setupRatioModelRight(t, false)
	before := m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "-")
	if m.cfg.Config.DetailPane.WidthRatio >= before {
		t.Errorf("- detail-focus must shrink detail: before=%.3f after=%.3f",
			before, m.cfg.Config.DetailPane.WidthRatio)
	}
	// list focus: `-` grows detail
	m = setupRatioModelRight(t, true)
	before = m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "-")
	if m.cfg.Config.DetailPane.WidthRatio <= before {
		t.Errorf("- list-focus must grow detail: before=%.3f after=%.3f",
			before, m.cfg.Config.DetailPane.WidthRatio)
	}
}

// TestModel_T155_Pipe_DetailFocus_Toggles_Right: `|` with detail focused
// toggles 0.30 ↔ 0.50.
func TestModel_T155_Pipe_DetailFocus_Toggles_Right(t *testing.T) {
	m := setupRatioModelRight(t, false)
	// From default 0.30 → 0.50.
	m = key(m, "|")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != 0.50 {
		t.Errorf("| detail from 0.30: got %.3f, want 0.50", got)
	}
	// Back to 0.30.
	m = key(m, "|")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != 0.30 {
		t.Errorf("| detail from 0.50: got %.3f, want 0.30", got)
	}
}

// TestModel_T155_Pipe_ListFocus_TogglesShare_Right: `|` with list
// focused toggles list share 0.30 ↔ 0.50 (detail 0.70 ↔ 0.50).
func TestModel_T155_Pipe_ListFocus_TogglesShare_Right(t *testing.T) {
	m := setupRatioModelRight(t, true)
	// Default detail=0.30 → list share=0.70 (off-preset) → first preset share=0.30 → detail=0.70.
	m = key(m, "|")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != 0.70 {
		t.Errorf("| list from detail=0.30: got %.3f, want 0.70", got)
	}
	// From detail=0.70 (share=0.30) → toggle to share=0.50 → detail=0.50.
	m = key(m, "|")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != 0.50 {
		t.Errorf("| list from detail=0.70: got %.3f, want 0.50", got)
	}
}

// TestModel_T155_Equals_ResetsDefault_BothFocus_Right: `=` resets detail
// to 0.30 regardless of focus.
func TestModel_T155_Equals_ResetsDefault_BothFocus_Right(t *testing.T) {
	for _, listFocus := range []bool{false, true} {
		m := setupRatioModelRight(t, listFocus)
		// Mutate away from default first.
		m.cfg.Config.DetailPane.WidthRatio = 0.50
		m.layout = m.layout.SetWidthRatio(0.50)
		m = key(m, "=")
		if got := m.cfg.Config.DetailPane.WidthRatio; got != appshell.RatioDefault {
			t.Errorf("= listFocus=%v: got %.3f, want %.3f",
				listFocus, got, appshell.RatioDefault)
		}
	}
}

// TestModel_T155_ClampPin_DetailFocus_Max_Right: `+` at RatioMax is a
// no-op with detail focused.
func TestModel_T155_ClampPin_DetailFocus_Max_Right(t *testing.T) {
	m := setupRatioModelRight(t, false)
	m.cfg.Config.DetailPane.WidthRatio = appshell.RatioMax
	m.layout = m.layout.SetWidthRatio(appshell.RatioMax)
	m = key(m, "+")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != appshell.RatioMax {
		t.Errorf("+ at RatioMax detail-focus: got %.3f, want %.3f (no-op)", got, appshell.RatioMax)
	}
}

// TestModel_T155_ClampPin_DetailFocus_Min_Right: `-` at RatioMin is a
// no-op with detail focused.
func TestModel_T155_ClampPin_DetailFocus_Min_Right(t *testing.T) {
	m := setupRatioModelRight(t, false)
	m.cfg.Config.DetailPane.WidthRatio = appshell.RatioMin
	m.layout = m.layout.SetWidthRatio(appshell.RatioMin)
	m = key(m, "-")
	if got := m.cfg.Config.DetailPane.WidthRatio; got != appshell.RatioMin {
		t.Errorf("- at RatioMin detail-focus: got %.3f, want %.3f (no-op)", got, appshell.RatioMin)
	}
}

// TestModel_T155_Below_ActivatesHeightRatio: below-mode mutates
// height_ratio (not width_ratio).
func TestModel_T155_Below_ActivatesHeightRatio(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	beforeH := m.cfg.Config.DetailPane.HeightRatio
	beforeW := m.cfg.Config.DetailPane.WidthRatio
	m = key(m, "+")
	if m.cfg.Config.DetailPane.HeightRatio == beforeH {
		t.Errorf("+ in below-mode must change height_ratio: %.3f unchanged", beforeH)
	}
	if m.cfg.Config.DetailPane.WidthRatio != beforeW {
		t.Errorf("+ in below-mode must NOT change width_ratio: %.3f → %.3f",
			beforeW, m.cfg.Config.DetailPane.WidthRatio)
	}
}

// TestModel_T155_PaneClosed_AllKeys_NoOp: all four ratio keys are silent
// no-ops when the detail pane is closed (no divider to move).
func TestModel_T155_PaneClosed_AllKeys_NoOp(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	m = m.SetEntries(makeEntries(3))
	if m.pane.IsOpen() {
		t.Fatal("precondition: pane should be closed")
	}
	beforeW := m.cfg.Config.DetailPane.WidthRatio
	beforeH := m.cfg.Config.DetailPane.HeightRatio

	for _, k := range []string{"+", "-", "=", "|"} {
		m = key(m, k)
		if m.cfg.Config.DetailPane.WidthRatio != beforeW {
			t.Errorf("%q with pane closed must not change width_ratio: %.3f → %.3f",
				k, beforeW, m.cfg.Config.DetailPane.WidthRatio)
		}
		if m.cfg.Config.DetailPane.HeightRatio != beforeH {
			t.Errorf("%q with pane closed must not change height_ratio: %.3f → %.3f",
				k, beforeH, m.cfg.Config.DetailPane.HeightRatio)
		}
	}
	// No disk write should have fired — config file must not exist.
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("pane-closed ratio keys must not trigger config save; file exists: %v", err)
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
	if !m.draggingDivider {
		t.Errorf("Press on below-mode divider row y=%d must start drag", dy)
	}
}

// TestModel_T156_BelowDrag_MotionUp_GrowsDetail verifies motion to a
// smaller y during an active drag pushes height_ratio upward (detail
// grows when divider moves up).
func TestModel_T156_BelowDrag_MotionUp_GrowsDetail(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	dy := belowDividerY(m)
	before := m.cfg.Config.DetailPane.HeightRatio

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 4, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	after := m.cfg.Config.DetailPane.HeightRatio
	if after <= before {
		t.Errorf("drag up should grow height_ratio: before=%.3f after=%.3f", before, after)
	}
}

// TestModel_T156_BelowDrag_MotionDown_ShrinksDetail: motion to a larger y
// shrinks detail when the divider moves down.
func TestModel_T156_BelowDrag_MotionDown_ShrinksDetail(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	// Seed a roomier starting ratio so a downward motion still fits inside clamp.
	m.cfg.Config.DetailPane.HeightRatio = 0.50
	m.paneHeight = m.paneHeight.SetRatio(0.50)
	m = m.relayout()
	dy := belowDividerY(m)
	before := m.cfg.Config.DetailPane.HeightRatio

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: 20, Y: dy + 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	after := m.cfg.Config.DetailPane.HeightRatio
	if after >= before {
		t.Errorf("drag down should shrink height_ratio: before=%.3f after=%.3f", before, after)
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
	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: want below, got %v", m.resize.Orientation())
	}
	dy := belowDividerY(m)

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	// Ensure config file does NOT exist yet — motion must not save.
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	if _, err := os.Stat(cfgPath); err == nil {
		t.Error("motion during drag must NOT write config; file exists already")
	}
	m = send(m, tea.MouseMsg{X: 20, Y: dy - 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	if m.draggingDivider {
		t.Error("Release must end drag session")
	}
	reloaded := config.Load(cfgPath)
	if reloaded.Config.DetailPane.HeightRatio != m.cfg.Config.DetailPane.HeightRatio {
		t.Errorf("disk height_ratio after release: got %.3f, want %.3f",
			reloaded.Config.DetailPane.HeightRatio, m.cfg.Config.DetailPane.HeightRatio)
	}
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
			if m.focus != startFocus {
				t.Errorf("drag changed focus: start=%v end=%v", startFocus, m.focus)
			}
		})
	}
}

// TestModel_T156_BelowDrag_ClampsAtMax verifies a drag that would push the
// divider past the top edge pins to RatioMax.
func TestModel_T156_BelowDrag_ClampsAtMax(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	dy := belowDividerY(m)

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: 20, Y: -100, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	if got := m.cfg.Config.DetailPane.HeightRatio; got != appshell.RatioMax {
		t.Errorf("extreme up drag: got %.3f, want RatioMax %.3f", got, appshell.RatioMax)
	}
}

// TestModel_T156_BelowDrag_ClampsAtMin verifies a drag that would push the
// divider past the bottom edge pins to RatioMin.
func TestModel_T156_BelowDrag_ClampsAtMin(t *testing.T) {
	m := setupRatioModelBelow(t, false)
	dy := belowDividerY(m)
	termH := m.resize.Height()

	m = send(m, tea.MouseMsg{X: 20, Y: dy, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: 20, Y: termH + 100, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	if got := m.cfg.Config.DetailPane.HeightRatio; got != appshell.RatioMin {
		t.Errorf("extreme down drag: got %.3f, want RatioMin %.3f", got, appshell.RatioMin)
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
		if m.pane.IsOpen() {
			t.Fatalf("precondition: pane must be closed at %dx%d", size.w, size.h)
		}
		m = send(m, tea.MouseMsg{X: 10, Y: 10, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
		if m.draggingDivider {
			t.Errorf("%dx%d press with pane closed must not start drag", size.w, size.h)
		}
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("pane-closed press must not save config; file exists: %v", err)
	}
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
			if m.focus != startFocus {
				t.Errorf("divider click changed focus: start=%v end=%v", startFocus, m.focus)
			}
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
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right, got %v", m.resize.Orientation())
	}
	beforeH := m.cfg.Config.DetailPane.HeightRatio
	dx := rightDividerX(m)

	m = send(m, tea.MouseMsg{X: dx, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: dx - 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	if m.cfg.Config.DetailPane.HeightRatio != beforeH {
		t.Errorf("right-mode drag must not mutate height_ratio: %.3f → %.3f",
			beforeH, m.cfg.Config.DetailPane.HeightRatio)
	}
	reloaded := config.Load(cfgPath)
	if reloaded.Config.DetailPane.WidthRatio != m.cfg.Config.DetailPane.WidthRatio {
		t.Errorf("disk width_ratio after right-drag release: got %.3f, want %.3f",
			reloaded.Config.DetailPane.WidthRatio, m.cfg.Config.DetailPane.WidthRatio)
	}
}

// ---------- T-158: single-owner click-row resolver (cavekit-entry-list R10) ----------

// clickAt emits a left-Press at (x, y) and returns the updated model.
func clickAt(m Model, x, y int) Model {
	return send(m, tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
}

// TestModel_T158_Click_FirstRow_SelectsRowZero_BelowMode: click at the
// first visible content row (terminal y=2) selects visible row 0 — NOT
// row 2 (the prior bug).
func TestModel_T158_Click_FirstRow_SelectsRowZero_BelowMode(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(20))
	if m.pane.IsOpen() {
		t.Fatalf("precondition: pane should be closed")
	}
	m = clickAt(m, 10, 2)
	if got := m.list.CursorPosition(); got != 1 {
		t.Errorf("click y=2 (first content row): CursorPosition = %d (1-based), want 1", got)
	}
}

// TestModel_T158_Click_SecondRow_SelectsRowOne_BelowMode: y=3 → row 1.
func TestModel_T158_Click_SecondRow_SelectsRowOne_BelowMode(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(20))
	m = clickAt(m, 10, 3)
	if got := m.list.CursorPosition(); got != 2 {
		t.Errorf("click y=3 (second content row): CursorPosition = %d, want 2", got)
	}
}

// TestModel_T158_Click_TopBorder_NoOp: click at y=1 (list top border) does
// not move the cursor.
func TestModel_T158_Click_TopBorder_NoOp(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(20))
	before := m.list.CursorPosition()
	m = clickAt(m, 10, 1)
	if got := m.list.CursorPosition(); got != before {
		t.Errorf("click on top border: CursorPosition = %d, want %d (unchanged)", got, before)
	}
}

// TestModel_T158_Click_Header_NoOp: y=0 (header) routes to ZoneHeader, not
// list — cursor does not move.
func TestModel_T158_Click_Header_NoOp(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(20))
	for i := 0; i < 5; i++ {
		m = key(m, "j") // advance cursor to a non-zero row
	}
	before := m.list.CursorPosition()
	m = clickAt(m, 10, 0)
	if got := m.list.CursorPosition(); got != before {
		t.Errorf("click on header: CursorPosition = %d, want %d (unchanged)", got, before)
	}
}

// TestModel_T158_Click_FirstRow_RightMode: same single-owner mapping applies
// in right-split. y=2 → row 0.
func TestModel_T158_Click_FirstRow_RightMode(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(20)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right, got %v", m.resize.Orientation())
	}
	m = clickAt(m, 10, 2)
	if got := m.list.CursorPosition(); got != 1 {
		t.Errorf("right-mode click y=2: CursorPosition = %d, want 1", got)
	}
}

// TestModel_T158_Click_FirstRow_BelowMode_PaneOpen: same mapping applies
// with the detail pane open (EntryListHeight shrinks but contentTopY stays 2).
func TestModel_T158_Click_FirstRow_BelowMode_PaneOpen(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(20)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	if m.resize.Orientation() != appshell.OrientationBelow {
		t.Fatalf("precondition: want below, got %v", m.resize.Orientation())
	}
	m = clickAt(m, 10, 2)
	if got := m.list.CursorPosition(); got != 1 {
		t.Errorf("below-mode pane-open click y=2: CursorPosition = %d, want 1", got)
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
	if got := m.list.CursorPosition(); got != 2 {
		t.Fatalf("first click y=3: CursorPosition = %d, want 2", got)
	}
	// Second click at the SAME y=3 within 500ms triggers double-click,
	// which emits OpenDetailPaneMsg. We verify the cmd return.
	updated, cmd := m.Update(tea.MouseMsg{X: 10, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	m = updated.(Model)
	if cmd == nil {
		// Not every list state emits a command here, but if double-click
		// fired, we should see OpenDetailPaneMsg.
		t.Fatal("second click at same y should emit a cmd for double-click")
	}
	msg := cmd()
	open, ok := msg.(entrylist.OpenDetailPaneMsg)
	if !ok {
		t.Fatalf("double-click cmd: want OpenDetailPaneMsg, got %T", msg)
	}
	if open.Entry.LineNumber != entries[1].LineNumber {
		t.Errorf("double-click opens wrong entry: got line %d, want %d",
			open.Entry.LineNumber, entries[1].LineNumber)
	}
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

	if got := m.list.CursorPosition(); got != before {
		t.Errorf("click on divider row (y=%d): CursorPosition = %d, want %d (unchanged)", dy, got, before)
	}
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
	if !m.pane.IsOpen() {
		t.Fatal("precondition: pane must be open")
	}
	if m.resize.Orientation() != appshell.OrientationRight {
		t.Fatalf("precondition: want right orientation, got %v", m.resize.Orientation())
	}
	dividerX := rightDividerX(m)
	startW := m.cfg.Config.DetailPane.WidthRatio

	// Begin a drag with a Press on the divider.
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !m.draggingDivider {
		t.Fatalf("precondition: Press at divider x=%d must start drag", dividerX)
	}

	// Shrink the terminal so the right-split detail width falls below
	// MinDetailWidth (30). Usable = w-5, detailW = usable * 0.30.
	// Using w=70 → usable=65 → detailW=19 < 30 → auto-close fires.
	m = resize(m, 70, 24)
	if m.pane.IsOpen() {
		t.Fatalf("precondition: pane must have auto-closed at w=70")
	}
	if m.draggingDivider {
		t.Errorf("auto-close must clear draggingDivider (belt-and-braces)")
	}

	// Any subsequent Motion + Release must be a no-op for ratio + config.
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	m = send(m, tea.MouseMsg{X: 20, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	if m.cfg.Config.DetailPane.WidthRatio != startW {
		t.Errorf("width_ratio mutated after mid-drag auto-close: start=%.3f end=%.3f",
			startW, m.cfg.Config.DetailPane.WidthRatio)
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("mid-drag auto-close must NOT write config; file exists: %v", err)
	}
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
	if m.draggingDivider {
		t.Error("closed-pane drag-branch guard must clear draggingDivider")
	}
	if m.cfg.Config.DetailPane.WidthRatio != startW {
		t.Errorf("closed-pane drag-branch must not mutate ratio: %.3f → %.3f",
			startW, m.cfg.Config.DetailPane.WidthRatio)
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("closed-pane drag-branch must not write config; file exists: %v", err)
	}
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
	if !m.draggingDivider {
		t.Fatalf("precondition: Press at divider x=%d must start drag", dividerX)
	}
	if m.dragDirty {
		t.Error("Press must reset dragDirty to false")
	}
	m = send(m, tea.MouseMsg{X: dividerX, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	if m.draggingDivider {
		t.Error("Release must end drag session")
	}
	if m.cfg.Config.DetailPane.WidthRatio != startW {
		t.Errorf("bare click mutated width_ratio: start=%.3f end=%.3f",
			startW, m.cfg.Config.DetailPane.WidthRatio)
	}
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("bare Press+Release on divider must NOT write config; file exists: %v", err)
	}
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
	if !m.dragDirty {
		t.Error("Motion that changes ratio must set dragDirty=true")
	}
	m = send(m, tea.MouseMsg{X: dividerX - 30, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("drag with motion must write config: %v", err)
	}
}

// ---------- F-132: degenerate-dim guard fidelity (supersedes T-165) ----------
//
// The original T-165 tests drove a 0-dim WindowSizeMsg which auto-closed
// the pane, then re-set draggingDivider=true and sent Motion. With pane
// closed, the !m.pane.IsOpen() guard at handleMouse model.go:524
// short-circuits BEFORE the termW/termH<=0 guard at :554-556 / :565-567
// is ever reached. Removing the degenerate-dim guards left the T-165
// tests green — proving the tests asserted behaviour via the wrong code
// path. /ck:review Pass 2 (F-132). The replacement tests below force the
// pane open at Motion time so the IsOpen() guard passes and the only
// thing standing between Motion and ratio-shadowing is the degenerate-dim
// guard. cavekit-app-shell.md R15 degenerate-dim AC was sharpened in the
// same /ck:revise --trace cycle to mandate this test shape.

// TestModel_F132_DegenerateDim_Right_GuardFiresWith_PaneOpen pins the
// right-split termW<=0 caller-guard at model.go:554-556. Forces 0-dim
// resize → pane auto-closes → re-opens pane via openPane (relayout does
// NOT re-trigger auto-close, so pane stays open with termW=0) → activates
// drag flag → sends Motion. If the guard is removed,
// RatioFromDragX(10, 0) returns ClampRatio(RatioDefault)=0.30 and shadows
// the persisted 0.55 — failing this test.
func TestModel_F132_DegenerateDim_Right_GuardFiresWith_PaneOpen(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "right"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 200, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m.cfg.Config.DetailPane.WidthRatio = 0.55
	m.layout = m.layout.SetWidthRatio(0.55)

	// Force termW=0 — auto-close fires, pane closes, draggingDivider clears.
	m = send(m, tea.WindowSizeMsg{Width: 0, Height: 24})
	if m.pane.IsOpen() {
		t.Fatalf("precondition: pane must auto-close at width=0")
	}
	// Re-open the pane. relayout() does not re-evaluate ShouldAutoCloseDetail
	// (that lives in the WindowSizeMsg handler only), so the pane stays open
	// even though m.resize.Width() is still 0.
	m = m.openPane(entries[0])
	if !m.pane.IsOpen() {
		t.Fatalf("precondition: pane must re-open via openPane")
	}
	if m.resize.Width() != 0 {
		t.Fatalf("precondition: termW must remain 0, got %d", m.resize.Width())
	}
	// All preconditions for the termW<=0 guard are now met: pane open,
	// drag flag set, Motion incoming with termW=0. The IsOpen() guard
	// passes — the only thing preventing ratio shadowing is :554-556.
	m.draggingDivider = true
	startRatio := m.cfg.Config.DetailPane.WidthRatio

	m = send(m, tea.MouseMsg{X: 10, Y: 5, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	if m.cfg.Config.DetailPane.WidthRatio != startRatio {
		t.Errorf("termW<=0 guard absent or bypassed: ratio shadowed from %.3f to %.3f",
			startRatio, m.cfg.Config.DetailPane.WidthRatio)
	}
}

// TestModel_F132_DegenerateDim_Below_GuardFiresWith_PaneOpen is the
// below-mode analog pinning model.go:565-567 (termH<=0 guard).
func TestModel_F132_DegenerateDim_Below_GuardFiresWith_PaneOpen(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/config.toml"
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "below"
	m := New("", false, cfgPath, cfg)
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m.cfg.Config.DetailPane.HeightRatio = 0.45
	m.paneHeight = m.paneHeight.SetRatio(0.45)

	m = send(m, tea.WindowSizeMsg{Width: 80, Height: 0})
	if m.pane.IsOpen() {
		t.Fatalf("precondition: pane must auto-close at height=0")
	}
	m = m.openPane(entries[0])
	if !m.pane.IsOpen() {
		t.Fatalf("precondition: pane must re-open via openPane")
	}
	if m.resize.Height() != 0 {
		t.Fatalf("precondition: termH must remain 0, got %d", m.resize.Height())
	}
	m.draggingDivider = true
	startRatio := m.cfg.Config.DetailPane.HeightRatio

	m = send(m, tea.MouseMsg{X: 20, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})

	if m.cfg.Config.DetailPane.HeightRatio != startRatio {
		t.Errorf("termH<=0 guard absent or bypassed: ratio shadowed from %.3f to %.3f",
			startRatio, m.cfg.Config.DetailPane.HeightRatio)
	}
}
