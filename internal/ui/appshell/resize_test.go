package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// T-053: R7.1 — handle tea.WindowSizeMsg, update dimensions.
func TestResizeModel_HandleWindowSizeMsg(t *testing.T) {
	m := NewResizeModel(80, 24)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m2.Width() != 120 || m2.Height() != 40 {
		t.Errorf("after resize: got %dx%d, want 120x40", m2.Width(), m2.Height())
	}
}

// T-053: R7.2 — layout is re-built preserving pane proportions.
func TestApplyToLayout_PreservesProportions(t *testing.T) {
	rm := NewResizeModel(80, 24)
	l1 := ApplyToLayout(rm, 0.30, true) // 30% of 24 = 7 rows

	// Now resize to 40 rows.
	rm2, _ := rm.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	l2 := ApplyToLayout(rm2, 0.30, true) // 30% of 40 = 12 rows

	// Pane height should scale.
	if l2.DetailPaneHeight <= l1.DetailPaneHeight {
		t.Errorf("detail pane should be taller after terminal grows: %d vs %d",
			l2.DetailPaneHeight, l1.DetailPaneHeight)
	}
}

// T-053: R7.3 — no clipping when detail pane is closed.
func TestApplyToLayout_NoCrashDetailPaneClosed(t *testing.T) {
	rm := NewResizeModel(80, 10)
	l := ApplyToLayout(rm, 0.30, false)
	if l.EntryListHeight() < 1 {
		t.Errorf("entry list height should be at least 1: got %d", l.EntryListHeight())
	}
}
