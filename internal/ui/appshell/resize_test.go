package appshell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
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
	if rm.Orientation() != OrientationRight {
		t.Fatalf("initial orientation at 120 cols: got %v, want right", rm.Orientation())
	}

	rm, _ = rm.Update(tea.WindowSizeMsg{Width: 90, Height: 30})
	if rm.Orientation() != OrientationBelow {
		t.Errorf("orientation after shrink to 90 cols: got %v, want below", rm.Orientation())
	}
	// ResizeModel must not mutate the cfg ratios on resize.
	if rm.cfg.DetailPane.HeightRatio != 0.55 {
		t.Errorf("height_ratio mutated by resize: got %.3f, want 0.550", rm.cfg.DetailPane.HeightRatio)
	}
	if rm.cfg.DetailPane.WidthRatio != 0.25 {
		t.Errorf("width_ratio mutated by resize: got %.3f, want 0.250", rm.cfg.DetailPane.WidthRatio)
	}

	// Flip back: orientation should re-evaluate on every WindowSizeMsg.
	rm, _ = rm.Update(tea.WindowSizeMsg{Width: 130, Height: 30})
	if rm.Orientation() != OrientationRight {
		t.Errorf("orientation after grow back to 130 cols: got %v, want right", rm.Orientation())
	}
	if rm.cfg.DetailPane.HeightRatio != 0.55 || rm.cfg.DetailPane.WidthRatio != 0.25 {
		t.Errorf("ratios mutated by second resize: height=%.3f width=%.3f", rm.cfg.DetailPane.HeightRatio, rm.cfg.DetailPane.WidthRatio)
	}
}
