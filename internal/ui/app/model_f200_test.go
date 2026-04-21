package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// colorANSIF200 renders a probe with color c and returns the SGR prefix
// lipgloss emits for it. Mirrors `colorANSI` in appshell/divider_test.go —
// kept local so assertions stay independent of termenv rounding.
func colorANSIF200(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
}

func firstRenderedRow(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// TestModel_F200_WindowSizeMsg_AutoFlip_RefreshesBelowMode verifies
// cavekit-app-shell.md R7 AC 9: when `detail_pane.position` is "auto" and
// the pane is open, a WindowSizeMsg that crosses `orientation_threshold_cols`
// must refresh the detail pane's below-mode rendering flag. Post-flip, the
// pane's top border row must match the NEW orientation's seam contract per
// R10 AC 10 — DragHandle in below-mode (pane sits below the list, so the
// top border IS the drag seam) and the focus-state color (not DragHandle)
// in right-mode (pane sits beside the list, so the top border is a regular
// pane border, not a seam).
//
// F-200 (Tier 23 /ck:check): `WithBelowMode` was wired in `relayout()` but
// omitted from the inline pane chain in the `WindowSizeMsg` handler. Auto
// flips left `pane.belowMode` stale and the rendered seam color drifted
// from the declared orientation. Must exercise BOTH flip directions
// (below→right and right→below) with the pane open throughout.
func TestModel_F200_WindowSizeMsg_AutoFlip_RefreshesBelowMode(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.DetailPane.Position = "auto"
	cfg.Config.DetailPane.HeightRatio = 0.30
	cfg.Config.DetailPane.WidthRatio = 0.40
	cfg.Config.DetailPane.OrientationThresholdCols = 100

	m := New("", false, "", cfg)
	// Start in BELOW mode (80 cols, under threshold=100).
	m = resize(m, 80, 24)
	entries := makeEntries(3)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: 80 cols should be below, got %v", m.resize.Orientation())
	require.Truef(t, m.pane.IsOpen(), "precondition: pane should be open after openPane")

	dragSGR := colorANSIF200(m.th.DragHandle)
	require.NotEmptyf(t, dragSGR, "empty DragHandle SGR — TrueColor profile not active")

	// Precondition: below-mode top border is the drag seam (DragHandle SGR).
	topRow := firstRenderedRow(m.pane.View())
	require.Containsf(t, topRow, dragSGR,
		"precondition: below-mode pane top border should carry DragHandle SGR %q; got %q",
		dragSGR, topRow)

	// Flip 1: below → right via WindowSizeMsg crossing orientation_threshold_cols.
	m = resize(m, 140, 24)
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"140 cols should flip to right, got %v", m.resize.Orientation())
	require.Truef(t, m.pane.IsOpen(), "pane should still be open after flip to right")

	// In right-mode the pane's top border is a regular pane border (NOT a
	// drag seam — the seam in right-mode is the separate 1-cell `│` divider
	// glyph between list and pane, not on the pane's own border). So the
	// pane's top row must NOT carry DragHandle SGR.
	topRow = firstRenderedRow(m.pane.View())
	assert.NotContainsf(t, topRow, dragSGR,
		"F-200: after below→right WindowSizeMsg flip, pane.belowMode is stale — "+
			"pane top border still carries DragHandle SGR %q\ntop row: %q",
		dragSGR, topRow)

	// Flip 2: right → below. Pane should reclaim the DragHandle top border.
	m = resize(m, 80, 24)
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"80 cols should flip back to below, got %v", m.resize.Orientation())
	require.Truef(t, m.pane.IsOpen(), "pane should still be open after flip back to below")

	topRow = firstRenderedRow(m.pane.View())
	assert.Containsf(t, topRow, dragSGR,
		"F-200: after right→below WindowSizeMsg flip, pane.belowMode is stale — "+
			"pane top border missing DragHandle SGR %q\ntop row: %q",
		dragSGR, topRow)
}
