package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/ui/entrylist"
)

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
