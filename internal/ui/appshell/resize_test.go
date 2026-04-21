package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
)

// T-053: R7.1 — handle tea.WindowSizeMsg, update dimensions.
func TestResizeModel_HandleWindowSizeMsg(t *testing.T) {
	m := NewResizeModel(80, 24)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	assert.Equalf(t, 120, m2.Width(), "after resize width")
	assert.Equalf(t, 40, m2.Height(), "after resize height")
}

// T-053: R7.2 — layout is re-built preserving pane proportions.
func TestApplyToLayout_PreservesProportions(t *testing.T) {
	rm := NewResizeModel(80, 24)
	l1 := ApplyToLayout(rm, 0.30, true) // 30% of 24 = 7 rows

	// Now resize to 40 rows.
	rm2, _ := rm.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	l2 := ApplyToLayout(rm2, 0.30, true) // 30% of 40 = 12 rows

	// Pane height should scale.
	assert.Greaterf(t, l2.DetailPaneHeight, l1.DetailPaneHeight,
		"detail pane should be taller after terminal grows: %d vs %d",
		l2.DetailPaneHeight, l1.DetailPaneHeight)
}

// T-053: R7.3 — no clipping when detail pane is closed.
func TestApplyToLayout_NoCrashDetailPaneClosed(t *testing.T) {
	rm := NewResizeModel(80, 10)
	l := ApplyToLayout(rm, 0.30, false)
	assert.GreaterOrEqualf(t, l.EntryListHeight(), 1, "entry list height should be at least 1: got %d", l.EntryListHeight())
}

// T-108: with position="auto", a shrink from 120 to 90 cols flips orientation
// from right → below, and both ratios in the attached config remain at the
// values they were configured with (no clobbering).
func TestResizeModel_AutoFlipPreservesBothRatios(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.DetailPane.Position = "auto"
	cfg.DetailPane.OrientationThresholdCols = 100
	cfg.DetailPane.HeightRatio = 0.55
	cfg.DetailPane.WidthRatio = 0.25

	rm := NewResizeModel(120, 30).WithConfig(cfg)
	require.Equalf(t, OrientationRight, rm.Orientation(), "initial orientation at 120 cols")

	rm, _ = rm.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	assert.Equalf(t, OrientationBelow, rm.Orientation(), "orientation after shrink to 90 cols")
	// ResizeModel must not mutate the cfg ratios on resize.
	assert.Equalf(t, 0.55, rm.cfg.DetailPane.HeightRatio, "height_ratio mutated by resize")
	assert.Equalf(t, 0.25, rm.cfg.DetailPane.WidthRatio, "width_ratio mutated by resize")

	// Flip back: orientation should re-evaluate on every WindowSizeMsg.
	rm, _ = rm.Update(tea.WindowSizeMsg{Width: 130, Height: 30})
	assert.Equalf(t, OrientationRight, rm.Orientation(), "orientation after grow back to 130 cols")
	assert.Equalf(t, 0.55, rm.cfg.DetailPane.HeightRatio, "height_ratio mutated by second resize")
	assert.Equalf(t, 0.25, rm.cfg.DetailPane.WidthRatio, "width_ratio mutated by second resize")
}
