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
// opens the detail pane and moves focus to it.
func TestModel_Enter_OpensDetailPane(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))

	m = key(m, "enter")

	if !m.pane.IsOpen() {
		t.Error("pane should be open after Enter")
	}
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("focus after Enter: got %v, want FocusDetailPane", m.focus)
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
	m = key(m, "enter") // open pane

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
	m = key(m, "enter") // open pane — focus is now FocusDetailPane

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
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("focus: got %v, want FocusDetailPane", m.focus)
	}
}

// ---------- T-096: Tab focus cycle ----------

// TestModel_Tab_CyclesListToDetails verifies Tab flips list → details when
// both panes are visible.
func TestModel_Tab_CyclesListToDetails(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // opens pane, focus → details
	// Move focus back to list via BlurredMsg shortcut for a clean start.
	m.focus = appshell.FocusEntryList
	m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)

	m = key(m, "tab")
	if m.focus != appshell.FocusDetailPane {
		t.Errorf("Tab from list with pane open: got %v, want FocusDetailPane", m.focus)
	}
	if !m.pane.IsOpen() {
		t.Error("Tab must not close the pane")
	}
}

// TestModel_Tab_WrapsDetailsToList verifies Tab wraps details → list.
func TestModel_Tab_WrapsDetailsToList(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "enter") // focus = FocusDetailPane

	m = key(m, "tab")
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
