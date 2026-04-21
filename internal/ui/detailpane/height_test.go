package detailpane

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// T-042: R6.1 — pane opens at configured height ratio.
func TestHeightModel_PaneHeight_Ratio(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	assert.Equal(t, 30, m.PaneHeight(), "PaneHeight")
}

// T-042: R6.2 — + increases height.
func TestHeightModel_PlusKey_Increases(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	assert.Greaterf(t, m2.PaneHeight(), m.PaneHeight(), "+ should increase pane height: before=%d after=%d", m.PaneHeight(), m2.PaneHeight())
}

// T-042: R6.3 — - decreases height.
func TestHeightModel_MinusKey_Decreases(t *testing.T) {
	m := NewHeightModel(0.50, 100)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	assert.Lessf(t, m2.PaneHeight(), m.PaneHeight(), "- should decrease pane height: before=%d after=%d", m.PaneHeight(), m2.PaneHeight())
}

// T-042: R6.4 — terminal resize maintains proportional height.
func TestHeightModel_TerminalResize_ProportionalHeight(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 200})
	// Ratio preserved, height doubles.
	assert.InDeltaf(t, 0.30, m2.Ratio(), 0.001, "ratio should be preserved after resize: got %.2f", m2.Ratio())
	assert.Equal(t, 60, m2.PaneHeight(), "PaneHeight after resize to 200")
}

// T-042: R6.5 — mouse drag via SetRatioFromDrag.
func TestHeightModel_DragResize(t *testing.T) {
	m := NewHeightModel(0.30, 100)
	m2 := m.SetRatioFromDrag(50)
	assert.Equal(t, 50, m2.PaneHeight(), "after drag to 50 rows")
}

// Ratio is clamped to [minHeightRatio, maxHeightRatio].
func TestHeightModel_Clamping(t *testing.T) {
	m := NewHeightModel(0.05, 100) // below min, clamped to 0.10
	assert.GreaterOrEqualf(t, m.Ratio(), minHeightRatio, "ratio should be clamped: got %.2f", m.Ratio())
	m2 := NewHeightModel(0.95, 100) // above max, clamped to 0.90
	assert.LessOrEqualf(t, m2.Ratio(), maxHeightRatio, "ratio should be clamped: got %.2f", m2.Ratio())
}
