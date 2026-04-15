package detailpane

import (
	"math"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// T-042: R6.1 — pane opens at configured height ratio.
func TestHeightModel_PaneHeight_Ratio(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	h := m.PaneHeight()
	if h != 30 {
		t.Errorf("PaneHeight: got %d, want 30", h)
	}
}

// T-042: R6.2 — + increases height.
func TestHeightModel_PlusKey_Increases(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if m2.PaneHeight() <= m.PaneHeight() {
		t.Errorf("+ should increase pane height: before=%d after=%d", m.PaneHeight(), m2.PaneHeight())
	}
}

// T-042: R6.3 — - decreases height.
func TestHeightModel_MinusKey_Decreases(t *testing.T) {
	m := NewHeightModel(0.50, 100)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if m2.PaneHeight() >= m.PaneHeight() {
		t.Errorf("- should decrease pane height: before=%d after=%d", m.PaneHeight(), m2.PaneHeight())
	}
}

// T-042: R6.4 — terminal resize maintains proportional height.
func TestHeightModel_TerminalResize_ProportionalHeight(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 200})
	// Ratio preserved, height doubles.
	gotRatio := m2.Ratio()
	if math.Abs(gotRatio-0.30) > 0.001 {
		t.Errorf("ratio should be preserved after resize: got %.2f", gotRatio)
	}
	if m2.PaneHeight() != 60 {
		t.Errorf("PaneHeight after resize to 200: got %d, want 60", m2.PaneHeight())
	}
}

// T-042: R6.5 — mouse drag via SetRatioFromDrag.
func TestHeightModel_DragResize(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2 := m.SetRatioFromDrag(50)
	if m2.PaneHeight() != 50 {
		t.Errorf("after drag to 50 rows: got %d, want 50", m2.PaneHeight())
	}
}

// Ratio is clamped to [minHeightRatio, maxHeightRatio].
func TestHeightModel_Clamping(t *testing.T) {
	m := NewHeightModel(0.05, 100) // below min, clamped to 0.10
	if m.Ratio() < minHeightRatio {
		t.Errorf("ratio should be clamped: got %.2f", m.Ratio())
	}
	m2 := NewHeightModel(0.95, 100) // above max, clamped to 0.90
	if m2.Ratio() > maxHeightRatio {
		t.Errorf("ratio should be clamped: got %.2f", m2.Ratio())
	}
}
