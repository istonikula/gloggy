package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

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
	assert.True(t, m.loading.IsActive(),
		"BUG: loading should be active when model is created with a file path, but is not")
}

// TestModel_StdinMode_LoadingNotActive verifies that stdin mode (empty sourceName)
// does not activate the loading indicator.
func TestModel_StdinMode_LoadingNotActive(t *testing.T) {
	m := newModel() // empty sourceName
	assert.False(t, m.loading.IsActive(), "loading should NOT be active for stdin mode")
}

// TestModel_FollowMode_LoadingNotActive verifies that follow/tail mode does not
// activate the loading indicator (tail is streaming, not batch loading).
func TestModel_FollowMode_LoadingNotActive(t *testing.T) {
	m := New("app.log", true, "", testCfg())
	assert.False(t, m.loading.IsActive(),
		"loading should NOT be active for follow mode (it is a streaming tail, not a batch load)")
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
	require.NotZero(t, renderedBefore, "expected non-zero rendered rows before opening pane")

	// Open the detail pane — this calls relayout(), which should shrink the list viewport.
	m = m.openPane(entries[0])

	renderedAfter := m.list.RenderedRowCount()

	// BUG: renderedAfter == renderedBefore because relayout discards the list update.
	// After the fix, the list viewport is reduced by the pane height, so renderedAfter < renderedBefore.
	assert.Lessf(t, renderedAfter, renderedBefore,
		"BUG: rendered row count did not decrease after pane opened "+
			"(relayout discards list update); before=%d after=%d",
		renderedBefore, renderedAfter)
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

	assert.Equalf(t, len(entries), m.visibleCount(),
		"visibleCount with no filter: got %d, want %d", m.visibleCount(), len(entries))
}

// ---------- quit ----------

// TestModel_Q_EmitsQuitCmd verifies the 'q' key emits a tea.Quit command.
func TestModel_Q_EmitsQuitCmd(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd, "expected non-nil cmd from 'q' key")
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.Truef(t, ok, "expected tea.QuitMsg, got %T", msg)
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

	assert.False(t, m.loading.IsActive(), "loading should be inactive after LoadDoneMsg")
}

// ---------- window resize ----------

// TestModel_WindowSizeMsg_UpdatesDimensions verifies that a WindowSizeMsg is
// propagated to the resize model so Width() and Height() reflect the new size.
func TestModel_WindowSizeMsg_UpdatesDimensions(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 40)

	assert.Equalf(t, 120, m.resize.Width(), "resize.Width(): got %d, want 120", m.resize.Width())
	assert.Equalf(t, 40, m.resize.Height(), "resize.Height(): got %d, want 40", m.resize.Height())
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
	assert.Truef(t, containsCount(headerView, 7),
		"header should show count 7 after SetEntries; got: %q", headerView)
}

// TestModel_OpenDetailPane_ViaMsg verifies OpenDetailPaneMsg (double-click path)
// opens the pane.
func TestModel_OpenDetailPane_ViaMsg(t *testing.T) {
	m := newModel()
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)

	m = send(m, entrylist.OpenDetailPaneMsg{Entry: entries[1]})

	assert.True(t, m.pane.IsOpen(), "pane should be open after OpenDetailPaneMsg")
	// T-126 (F-017): opening the pane does NOT transfer focus.
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus: got %v, want FocusEntryList (pane open does not steal focus)", m.focus)
}
